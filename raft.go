package bigger

import (
	"fmt"
	"io"
	"io/ioutil"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
	"errors"
	"net"
	"encoding/gob"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"github.com/gobigger/bigger/raft"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second

	raftCmdRead		= "READ"
	raftCmdWrite	= "WRITE"
	raftCmdDelete	= "DELETE"
)

var (
	raftErrorShutdown		= errors.New("Store was shutdown")
	raftErrorAlreadyOpened	= errors.New("Store was already opened")
	raftErrorKeyNotFound	= errors.New("Key not present in store")
)


type (
	raftLogger struct {}
)

func (log *raftLogger) Write(p []byte) (n int, err error) {
	if p == nil {
		return 0, errors.New("no write data")
	}
	s := string(p)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}

	Bigger.Info("[raft]", s)
	return len(p), nil
}



// Store is a simple key-value store, where all changes are made via Raft consensus.
type raftStore struct {
	Bind	string
	Data	string
	File	string

	authMethodPubKey ssh.AuthMethod
	checkHostKey     func(addr string, remote net.Addr, key ssh.PublicKey) error
	sshTransport     *sshTransport
	raftTransport    *raft.NetworkTransport
	logStore         *raft.BuntStore

	mutex	sync.Mutex
	store	map[string][]byte // The key-value store for the system.
	opened	bool

	raft *raft.Raft // The consensus mechanism

	logger *log.Logger
}

// New returns a new Store.
// If debug is true, informational and debug messages are printed to os.Stderr
func newRaftStore(dir, bind string) *raftStore {
	return &raftStore{
		Data: dir, Bind: bind, File: "raft.db",
		store:      make(map[string][]byte),
		logger: log.New(&raftLogger{}, "", 0),
	}
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
func (s *raftStore) Open(join string) error {

	if s.opened {
		return raftErrorAlreadyOpened
	}
	s.opened = true

	var err error

	//TODO add error return to newSSHTransport
	s.sshTransport, s.raftTransport, err = newSSHTransport(s.Bind, s.Data, s.logger)
	if err != nil {
		// s.logger.Println("Error initializing ssh transport:", err)
		return fmt.Errorf("[raft]Transport初始化失败: %s", err)
	}

	s.authMethodPubKey = ssh.PublicKeys(s.sshTransport.PrivateKey)
	s.checkHostKey = s.sshTransport.CheckHostKey

	// Create peer storage.
	peerStore := raft.NewJSONPeers(s.Data, s.raftTransport)

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.Data, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("[raft]Snapshot加载失败: %s", err)
	}

	// Create the log store and stable store.
	s.logStore, err = raft.NewBuntStore(filepath.Join(s.Data, s.File))
	if err != nil {
		return fmt.Errorf("[raft]LogStore加载失败: %s", err)
	}


	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.Logger = s.logger

	peers, err := peerStore.Peers()
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if (len(peers) <= 1 && join != "") == false {
		config.EnableSingleNode = true
		config.DisableBootstrapAfterElect = false
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*raftFSM)(s), s.logStore, s.logStore, snapshots, peerStore, s.raftTransport)
	if err != nil {
		return fmt.Errorf("[raft]Raft创建失败: %s", err)
	}
	s.raft = ra




	go func() {

		for {
			joinMessage, notClosed := <-s.sshTransport.JoinMessage

			if !notClosed {
				return
			}

			if s.raft.State() != raft.Leader {
				//member转发join到leader

				s.logger.Printf("[%v] 转发加入消息至领导者", s.Bind)

				err := s.leaderJoin(joinMessage.JoinAddr)
				if err != nil {
					// s.logger.Println("Join转发至Leader失败：", joinMessage)
					joinMessage.ReturnChan <- false
				} else {
					joinMessage.ReturnChan <- true
				}

				close(joinMessage.ReturnChan)
				continue
			}

			//leader处理join消息
			err := s.join(joinMessage.JoinAddr)
			if err != nil {
				joinMessage.ReturnChan <- false
			} else {
				joinMessage.ReturnChan <- true
			}

			// Bigger.Trace("[raft]", "Join消息处理", err)

			close(joinMessage.ReturnChan)
		}

	}()

	go func() {

		for {
			leaderMessage, notClosed := <-s.sshTransport.LeaderMessage

			if !notClosed {
				return
			}

			if s.raft.State() != raft.Leader {
				//member转发request到leader
				err := s.leaderRequest(leaderMessage.Cmd)
				if err != nil {
					leaderMessage.ReturnChan <- false
				} else {
					leaderMessage.ReturnChan <- true
				}

				Bigger.Trace("[raft]", "Member转发Request到Reader", *leaderMessage.Cmd, err)

				close(leaderMessage.ReturnChan)
				continue
			}



			c := leaderMessage.Cmd
			sc, err := sshSerializeCommand(c)
			if err != nil {
				leaderMessage.ReturnChan <- false
				close(leaderMessage.ReturnChan)
				continue
			}

			f := s.raft.Apply(sc, raftTimeout)
			if _, ok := f.(error); ok {
				// s.logger.Println("Error applying command in leader request:", *leaderMessage.Cmd, err)
				leaderMessage.ReturnChan <- false
				close(leaderMessage.ReturnChan)
				continue
			}

			if f.Error() != nil {
				// s.logger.Println("Error distributing command in leader request:", *leaderMessage.Cmd, err)
				leaderMessage.ReturnChan <- false
				close(leaderMessage.ReturnChan)
				continue
			}

			leaderMessage.ReturnChan <- true
			close(leaderMessage.ReturnChan)
		}

	}()


	//是否加入节点
	if config.EnableSingleNode==false && join != "" {
		if err := s.Join(join, s.Bind); err != nil {
			return fmt.Errorf("[raft] 加入节点失败 [%s]", join)
		}
	}

	return nil
}

// Close closes the store after stepping down as node/leader.
func (s *raftStore) Close() error {

	shutdownFuture := s.raft.Shutdown()

	if err := shutdownFuture.Error(); err != nil {
		// s.logger.Println("raft shutdown error:", err)
		return err
	}

	if err := s.logStore.Close(); err != nil {
		// s.logger.Println("raftboltdb shutdown error:", err)
		return err
	}

	if err := s.raftTransport.Close(); err != nil {
		// s.logger.Println("raft transport close error:", err)
		return err
	}

	// s.logger.Println("successfully shutdown")
	return nil

}

// Join joins a node reachable under raftAddr, to the cluster lead by the
// node reachable under joinAddr. The joined node must be ready to respond to Raft
// communications at that raftAddr.
func (s *raftStore) Join(joinAddr, raftAddr string) error {

	if err := s.checkState(); err != nil {
		return err
	}

	sshClientConfig := &ssh.ClientConfig{
		User: raft.SSHProtocolUser,
		Auth: []ssh.AuthMethod{
			s.authMethodPubKey,
		},
		HostKeyCallback: s.checkHostKey,
	}

	serverConn, err := ssh.Dial("tcp", joinAddr, sshClientConfig)
	if err != nil {
		// s.logger.Printf("Server dial error: %s\n", err)
		return err
	}

	reply, _, err := serverConn.SendRequest(raft.SSHJoinRequestType, true, []byte(raftAddr))

	if err != nil {
		// s.logger.Println("Error sending out-of-band join request:", err)
		return err
	}

	if reply != true {
		// s.logger.Printf("Error adding peer on join node %s: %s\n", joinAddr, err)
		return err
	}

	return nil

}

func (s *raftStore) State() raft.RaftState {
	return s.raft.State()
}
func (s *raftStore) checkState() error {
	if s.raft.State() == raft.Shutdown {
		return raftErrorShutdown
	}
	return nil
}

//转发join到leader
func (s *raftStore) leaderJoin(raftAddr string) error {

	if err := s.checkState(); err != nil {
		return err
	}

	sshClientConfig := &ssh.ClientConfig{
		User: raft.SSHProtocolUser,
		Auth: []ssh.AuthMethod{
			s.authMethodPubKey,
		},
		HostKeyCallback: s.checkHostKey,
	}

	serverConn, err := ssh.Dial("tcp", s.raft.Leader(), sshClientConfig)
	if err != nil {
		// s.logger.Printf("Server dial error: %s\n", err)
		return err
	}

	reply, _, err := serverConn.SendRequest(raft.SSHJoinRequestType, true, []byte(raftAddr))

	if err != nil {
		// s.logger.Println("Error sending out-of-band join request:", err)
		return err
	}

	if reply != true {
		// s.logger.Printf("Error adding peer on join node %s: %s\n", raftAddr, err)
		return err
	}

	return nil

}

func (s *raftStore) leaderRequest(op *sshCommand) error {
	if err := s.checkState(); err != nil {
		return err
	}

	sshClientConfig := &ssh.ClientConfig{
		User: raft.SSHProtocolUser,
		Auth: []ssh.AuthMethod{
			s.authMethodPubKey,
		},
		HostKeyCallback: s.checkHostKey,
	}

	serverConn, err := ssh.Dial("tcp", s.raft.Leader(), sshClientConfig)
	if err != nil {
		// s.logger.Printf("Server dial error: %s\n", err)
		return err
	}

	sc, err := sshSerializeCommand(op)
	if err != nil {
		// s.logger.Printf("Command serialization error: %s\n", err)
		return err
	}

	reply, _, err := serverConn.SendRequest(raft.SSHLeaderMessageType, true, sc)

	if err != nil {
		// s.logger.Println("Error sending out-of-band leader request:", err)
		return err
	}

	if reply != true {
		// s.logger.Printf("Error executing command on leader node %s: %s\n", s.raft.Leader(), err)
		return err
	}

	return nil

}


func (s *raftStore) Read(key string) ([]byte, error) {
	if err := s.checkState(); err != nil {
		return []byte{}, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if val, ok := s.store[key]; ok {
		return val, nil 	   
	}

	return []byte{}, raftErrorKeyNotFound
}

// Set sets the value for the given key.
func (s *raftStore) Write(key string, value []byte) error {
	if err := s.checkState(); err != nil {
		return err
	}

	c := &sshCommand{
		Cmd:	raftCmdWrite,
		Key:	key,
		Value:	value,
	}

	if s.raft.State() != raft.Leader {
		// s.logger.Println("Forwarding Set command to leader: ", s.raft.Leader())
		return s.leaderRequest(c)
	}

	sc, err := sshSerializeCommand(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(sc, raftTimeout)
	if err, ok := f.(error); ok {
		return err
	}

	return f.Error()
}

// Delete deletes the given key.
func (s *raftStore) Delete(key string) error {
	if err := s.checkState(); err != nil {
		return err
	}

	c := &sshCommand{
		Cmd:	raftCmdDelete,
		Key:	key,
	}

	if s.raft.State() != raft.Leader {
		// s.logger.Println("Forwarding Delete command to leader: ", s.raft.Leader())
		return s.leaderRequest(c)
	}

	sc, err := sshSerializeCommand(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(sc, raftTimeout)
	if err, ok := f.(error); ok {
		return err
	}

	return f.Error()
}

func (s *raftStore) join(addr string) error {
	// s.logger.Printf("received join request for remote node as %s", addr)

	f := s.raft.AddPeer(addr)
	if f.Error() != nil {
		return f.Error()
	}
	// Bigger.Info("[raft]", "节点", addr, "加入成功")
	s.logger.Printf("节点 [%v] 加入成功", addr)
	return nil
}








type raftFSM raftStore

// Apply applies a Raft log entry to the key-value store.
func (f *raftFSM) Apply(l *raft.Log) interface{} {
	dsc, err := raft.SSHDeserializeCommand(l.Data)
	// fatalLogger := log.New(os.Stderr, "[raftd]", log.LstdFlags)

	if err != nil {
		//TODO fix fatal for library
		// fatalLogger.Fatalf("error in deserializeCommand: %s\n", err)
		Bigger.Warning("[raft]", "解析命令失败", err)
	}

	switch dsc.Cmd {
	case raftCmdWrite:
		return f.applyWrite(dsc.Key, dsc.Value)
	case raftCmdDelete:
		return f.applyDelete(dsc.Key)
	default:
		//TODO fix fatal for library
		// fatalLogger.Fatalf("unrecognized command op: %s", dsc.Cmd)
		Bigger.Warning("[raft]", "不支持的命令", dsc.Cmd)
		return nil
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *raftFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Clone the map.
	o := make(map[string][]byte)
	for k, v := range f.store {
		o[k] = v
	}
	return &raftSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (f *raftFSM) Restore(rc io.ReadCloser) error {
	o := make(map[string][]byte)

	decoder := gob.NewDecoder(rc)
	err := decoder.Decode(&o)

	if err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.store = o
	return nil
}

func (f *raftFSM) applyWrite(key string, value []byte) interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.store[key] = value
	return nil
}

func (f *raftFSM) applyDelete(key string) interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.store, key)
	return nil
}





//fsmSnapshot
type raftSnapshot struct {
	store map[string][]byte
}

func (f *raftSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		buf := bytes.NewBuffer([]byte{})
		encoder := gob.NewEncoder(buf)

		err := encoder.Encode(f.store)
		if err != nil {
			return err
		}

		var n int
		// Write data to sink.
		if n, err = sink.Write(buf.Bytes()); err != nil {
			return err
		}

		if n != buf.Len() {
			return fmt.Errorf("[raft]快照未完成写入")
		}

		// Close the sink.
		if err := sink.Close(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f *raftSnapshot) Release() {
	//TODO snapshot release function
}











type sshTransport struct {
	peerPubkeys		*sshPeerPublicKeys
	JoinMessage		chan sshJoinMessage
	LeaderMessage	chan sshLeaderMessage
	PrivateKey		ssh.Signer
	logger			*log.Logger

	privateFile		string
	publicFile		string
}


type sshPeerPublicKeys struct {
	sync.RWMutex
	pubkeys []ssh.PublicKey // may or may not contain this nodes pubkey
}

const (
	sshBogusAddress string = "127.0.0.1:0"
	sshMaxPoolConnections        = 5
	sshConnectionTimeout         = 10 * time.Second
	sshProtocolUser       	string = "raft"
	sshJoinRequestType    		string = "joinRequest"
	sshLeaderMessageType  		string = "sshLeaderMessage"
)

type sshCommand struct {
	Cmd		string
	Key		string
	Value	[]byte
}

type sshJoinMessage struct {
	JoinAddr   string
	ReturnChan chan bool
}

type sshLeaderMessage struct {
	Cmd        *sshCommand
	ReturnChan chan bool
}



var (
	noAuthorizedPeers  = errors.New("No authorized peers file")
)


func newSSHTransport(bindAddr string, raftDir string, logger *log.Logger) (*sshTransport, *raft.NetworkTransport, error) {

	s := new(sshTransport)
	s.peerPubkeys = new(sshPeerPublicKeys)
	s.logger = logger
	s.publicFile = "raft.rsa"
	s.privateFile = "raft.key"

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PublicKeyCallback: s.keyAuth,
	}

	privateBytes, err := ioutil.ReadFile(filepath.Join(raftDir, s.privateFile))
	if err != nil {
		// logger.Println("Failed to load private key, trying to generate a new pair")
		s.logger.Println("正在生成私钥")
		privateBytes, err = s.generateSSHKey(raftDir)

		if err != nil {
			//No usable SSH private key obtained
			return nil, nil, err
		}

	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		// logger.Println("Failed to parse private key:", err)
		// Bigger.Debug("[raft]", "私钥解析失败", err)
		s.logger.Println("私钥解析失败")
		return nil, nil, err
	}

	// Bigger.Debug("[raft]", "节点公钥是", string(ssh.MarshalAuthorizedKey(private.PublicKey())))
	// logger.Println("Node public key is: ", string(ssh.MarshalAuthorizedKey(private.PublicKey())))

	s.PrivateKey = private
	config.AddHostKey(private)

	publicKeys, err := s.readAuthorizedPeerKeys((filepath.Join(raftDir, s.publicFile)))

	if err != nil && err != noAuthorizedPeers {
		s.logger.Println("无效公钥")
		// logger.Println("Error reading authorized peer keys in newSSHTransport:", err)
		return nil, nil, err
	}

	if err == noAuthorizedPeers || len(publicKeys) < 1 {

		err := ioutil.WriteFile((filepath.Join(raftDir, s.publicFile)), ssh.MarshalAuthorizedKey(private.PublicKey()), 0644)

		if err != nil {
			// logger.Println("No public keys and error writing out new authorized key file:", err)
			// s.logger.Println("无效公钥且文件写入失败")
			return nil, nil, err
		}

		// logger.Printf("Written out initial '%s', copy this key to other nodes to initialize keys\n", filepath.Join(raftDir, s.publicFile))

	}

	// logger.Println("Parsed pubkeys", publicKeys)

	s.peerPubkeys.Lock()
	s.peerPubkeys.pubkeys = append(s.peerPubkeys.pubkeys, publicKeys...)
	s.peerPubkeys.Unlock()

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", bindAddr)

	if err != nil {
		// logger.Println("failed to listen for connection on", bindAddr, ":", err)
		// Bigger.Debug("[raft]", "监听失败", bindAddr, err)
		s.logger.Printf("监听 [%v] 失败", bindAddr)
		return nil, nil, err
	}

	sshClientConfig := &ssh.ClientConfig{
		User: sshProtocolUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(private),
		},
		HostKeyCallback: s.CheckHostKey,
	}

	raftListener := &sshStreamLayer{
		sshListener:  listener,
		incoming:     make(chan sshConn, 15),
		clientConfig: sshClientConfig,
		// logger:       s.logger,
	}

	s.JoinMessage = make(chan sshJoinMessage)
	s.LeaderMessage = make(chan sshLeaderMessage)

	go func() {

		for {

			nConn, err := listener.Accept()

			if err != nil {
				// log.Println("failed to accept incoming connection, assuming closed listener and stopping goroutine: ", err)
				return
			}

			go func() {

				// Before use, a handshake must be performed on the incoming
				// net.Conn.
				sshConnection, chans, reqs, err := ssh.NewServerConn(nConn, config)
				if err != nil {
					// logger.Println("Failed to handshake:", err)
					nConn.Close()
					return
				}
				// The incoming Request channel must be serviced.
				go s.handleRequests(s.JoinMessage, s.LeaderMessage, reqs)

				// Service the incoming Channel channel.
				for newChannel := range chans {

					if newChannel.ChannelType() != "direct-tcpip" {
						newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
						continue
					}

					channel, requests, err := newChannel.Accept()
					if err != nil {
						// logger.Println("Could not accept channel:", err)
						continue
					}

					go ssh.DiscardRequests(requests)

					raftListener.incoming <- sshConn{
						channel,
						sshConnection.LocalAddr(),
						sshConnection.RemoteAddr(),
					}

				}
			}()

		}

	}()

	return s, raft.NewNetworkTransportWithLogger(raftListener, sshMaxPoolConnections, sshConnectionTimeout, logger), nil

}

func (transport *sshTransport) readAuthorizedPeerKeys(path string) (pubs []ssh.PublicKey, err error) {

	//TODO Read comment and determine valid peer addresses?

	bytesRead, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return
	} else if err != nil && os.IsNotExist(err) {
		err = noAuthorizedPeers
		return
	}

	if len(bytesRead) == 0 {
		return
	}

	var rest int = len(bytesRead)

	for rest > 0 {

		var pubkey ssh.PublicKey
		pubkey, _, _, bytesRead, err = ssh.ParseAuthorizedKey(bytesRead)

		if err != nil {
			// transport.logger.Println("error parsing ssh publickey from authorized peers:", err)
			return
		}

		pubs = append(pubs, pubkey)
		rest = len(bytesRead)

	}

	return

}

func (transport *sshTransport) handleRequests(joinChannel chan sshJoinMessage, sshLeaderMessageChan chan sshLeaderMessage, reqs <-chan *ssh.Request) {

	for req := range reqs {
		// transport.logger.Printf("Received out-of-band request: %+v", req)
		if req.Type == sshJoinRequestType {

			returnChan := make(chan bool)
			msg := sshJoinMessage{JoinAddr: string(req.Payload), ReturnChan: returnChan}
			joinChannel <- msg

			timeout := time.After(15 * time.Second)
			select {
			case response := <-returnChan:
				err := req.Reply(response, req.Payload)
				if err != nil {
					// transport.logger.Println("Error replying to join request for:", string(req.Payload))
				}
			case <-timeout:
				// transport.logger.Println("Timed out processing join request for:", string(req.Payload))
				err := req.Reply(false, []byte{})
				if err != nil {
					// transport.logger.Println("Error replying to join request for:", string(req.Payload))
				}
			}

			continue

		}

		if req.Type == sshLeaderMessageType {

			returnChan := make(chan bool)

			//Decode payload

			cmd, err := sshDeserializeCommand(req.Payload)

			if err != nil {
				// transport.logger.Println("Error deserializing payload:", err)
				err := req.Reply(false, []byte{})
				if err != nil {
					// transport.logger.Println("Error replying to leader request for:", string(req.Payload))
				}
			}

			msg := sshLeaderMessage{Cmd: cmd, ReturnChan: returnChan}
			sshLeaderMessageChan <- msg

			timeout := time.After(sshConnectionTimeout)
			select {
			case response := <-returnChan:
				err := req.Reply(response, []byte{})
				if err != nil {
					// transport.logger.Println("Error replying to leader request for:", cmd)
				}
			case <-timeout:
				// transport.logger.Println("Timed out processing leader request for:", cmd)
				err := req.Reply(false, []byte{})
				if err != nil {
					// transport.logger.Println("Error replying to leader request for:", cmd)
				}
			}

			continue

		}

		// transport.logger.Printf("Did not handle out of band request: %+v", req)
	}
}

func (transport *sshTransport) keyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

	// transport.logger.Println(conn.RemoteAddr(), "authenticate with", key.Type(), "for user", conn.User())

	if conn.User() != sshProtocolUser {
		return nil, errors.New("Wrong user for protocol offered by server")
	}

	transport.peerPubkeys.RLock()
	defer transport.peerPubkeys.RUnlock()

	for _, storedKey := range transport.peerPubkeys.pubkeys {

		if subtle.ConstantTimeCompare(key.Marshal(), storedKey.Marshal()) == 1 {
			return nil, nil
		}

	}

	return nil, errors.New("Public key not found")
}

func (transport *sshTransport) CheckHostKey(addr string, remote net.Addr, key ssh.PublicKey) error {

	//TODO check addr

	transport.peerPubkeys.RLock()
	defer transport.peerPubkeys.RUnlock()

	for _, storedKey := range transport.peerPubkeys.pubkeys {

		if subtle.ConstantTimeCompare(key.Marshal(), storedKey.Marshal()) == 1 {
			return nil
		}

	}

	return errors.New("Public key not found")

}

func (transport *sshTransport) generateSSHKey(targetDir string) (privateKeyPem []byte, err error) {

	//generate 4096 bit rsa keypair
	var privateKey *rsa.PrivateKey
	privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		// transport.logger.Println("error generating private key:", err)
		return
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}

	privateKeyPem = pem.EncodeToMemory(&privateKeyBlock)

	if len(targetDir) > 0 {
		//persist key to raftDir
		err = ioutil.WriteFile(filepath.Join(targetDir, transport.privateFile), privateKeyPem, 0600)

		if err != nil {
			// transport.logger.Println("rrror persisting generated ssh private key:", err)
		}
	}

	return

}

type sshStreamLayer struct {
	sshListener  net.Listener
	incoming     chan sshConn
	clientConfig *ssh.ClientConfig
	logger       *log.Logger
}

func (listener *sshStreamLayer) Accept() (net.Conn, error) {

	select {
	case l := <-listener.incoming:
		wrapper := &sshConn{l, l.localAddr, l.remoteAddr}
		return wrapper, nil
	}

}

func (listener *sshStreamLayer) Close() error {
	return listener.sshListener.Close()
}

func (listener *sshStreamLayer) Addr() net.Addr {
	return listener.sshListener.Addr()
}

func (listener *sshStreamLayer) Dial(address string, timeout time.Duration) (net.Conn, error) {

	serverConn, err := ssh.Dial("tcp", address, listener.clientConfig)
	if err != nil {
		// log.Printf("Server dial error: %s\n", err)
		return nil, err
	}

	//client address given here is bogus and ignored by server
	remoteConn, err := serverConn.Dial("tcp", sshBogusAddress)
	if err != nil {
		// log.Printf("Remote dial error: %s\n", err)
		return nil, err
	}

	return remoteConn, nil

}

type sshConn struct {
	ssh.Channel
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (wrapper *sshConn) LocalAddr() net.Addr {
	return wrapper.localAddr
}

func (wrapper *sshConn) RemoteAddr() net.Addr {
	return wrapper.remoteAddr
}

//TODO IO timeout operations support
func (wrapper *sshConn) SetDeadline(t time.Time) error {
	return nil
}

func (wrapper *sshConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (wrapper *sshConn) SetWriteDeadline(t time.Time) error {
	return nil
}






func sshSerializeCommand(c *sshCommand) ([]byte, error) {

	buf := bytes.NewBuffer([]byte{})

	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(c)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func sshDeserializeCommand(sc []byte) (*sshCommand, error) {

	if len(sc) < 1 {
		return nil, fmt.Errorf("Zero length serialization passed")
	}

	buf := bytes.NewBuffer(sc)

	decoder := gob.NewDecoder(buf)

	command := &sshCommand{}

	err := decoder.Decode(command)

	if err != nil {
		return nil, err
	}

	return command, nil

}

