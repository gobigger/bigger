package bigger



type (

	//根配置
	configConfig struct {
        Name    string  `toml:"name"`
		Mode    string  `toml:"mode"`
        Charset	string	`toml:"charset"`
        Domain	string	`toml:"domain"`
        
        Node    nodeConfig              `toml:"node"`

        Lang	map[string]langConfig   `toml:"lang"`
        
        Logger  LoggerConfig            `toml:"logger"`
        Mutex   MutexConfig             `toml:"mutex"`
        Session SessionConfig           `toml:"session"`

        File    map[string]FileConfig    `toml:"file"`
        Cache   map[string]CacheConfig  `toml:"cache"`
        Data    map[string]DataConfig   `toml:"data"`

        Plan    PlanConfig              `toml:"plan"`
        Event   map[string]EventConfig  `toml:"event"`
        Queue   map[string]QueueConfig  `toml:"queue"`
        Socket   map[string]SocketConfig  `toml:"socket"`

        Http    HttpConfig               `toml:"http"`
        Site    map[string]SiteConfig    `toml:"site"`
        View    ViewConfig               `toml:"view"`

        Path    PathConfig               `toml:"path"`

        setting Map		                `toml:"setting"`
    }
    

    nodeConfig struct {
        Id      int64   `toml:"id"`
        Name    string  `toml:"name"`
        Bind    string  `toml:"bind"`
        Join    string  `toml:"join"`
        Port    int     `toml:"port"`
    }


    //语言配置
    langConfig struct {
        Name    string      `toml:"name"`
        Text    string      `toml:"text"`
        Accepts	[]string	`toml:"accepts"`
    }




    LoggerConfig struct {
        Driver  string  `toml:"driver"`
        Flag    string  `toml:"flag"`
        Console bool    `toml:"console"`
        Level   string  `toml:"level"`
        Format  string  `toml:"format"`
        Setting Map     `toml:"setting"`
    }


    MutexConfig struct {
        Driver  string          `toml:"driver"`
        Expiry  string          `toml:"expiry"`
        Prefix  string          `toml:"prefix"`
        Setting Map             `toml:"setting"`
    }


    SessionConfig struct {
        Driver  string          `toml:"driver"`
        Expiry  string          `toml:"expiry"`
        Prefix  string          `toml:"prefix"`
        Setting Map             `toml:"setting"`
    }


    FileConfig struct {
        Driver  string  `toml:"driver"`
        Weight  int     `toml:"weight"`
        Expiry  string  `toml:"expiry"`
        Browse  string  `toml:"browse"`
        Preview string  `toml:"preview"`
        Setting Map     `toml:"setting"`
    }

    CacheConfig struct {
        Driver  string  `toml:"driver"`
        Expiry  string  `toml:"expiry"`
        Prefix  string  `toml:"prefix"`
        Setting Map     `toml:"setting"`
    }

    DataConfig struct {
        Driver  string  `toml:"driver"`
        Cache   string  `toml:"cache"`
        Url     string  `toml:"url"`
        Serial  string  `toml:"serial"`
        Setting Map     `toml:"setting"`
    }


    PlanConfig struct {
        Driver  string              `toml:"driver"`
        Setting Map                 `toml:"setting"`
        Prefix  string              `toml:"prefix"`
        Timer   map[string][]string `toml:"timer"`
    }
    EventConfig struct {
        Driver  string  `toml:"driver"`
        Weight  int     `toml:"weight"`
        Prefix  string  `toml:"prefix"`
        Setting Map     `toml:"setting"`
    }
    QueueConfig struct {
        Driver  string          `toml:"driver"`
        Weight  int             `toml:"weight"`
        Prefix  string          `toml:"prefix"`
        Setting Map             `toml:"setting"`
        Liner   map[string]int  `toml:"liner"`
    }

    SocketConfig struct {
        Driver  string  `toml:"driver"`
        Weight  int     `toml:"weight"`
        Prefix  string  `toml:"prefix"`
        Setting Map     `toml:"setting"`
    }

    HttpConfig struct {
        Driver	string  `toml:"driver"`
        Port    int     `toml:"port"`
        Expiry	string  `toml:"expiry"`
        MaxAge	string  `toml:"maxage"`
        Setting Map     `toml:"setting"`
    }

    SiteConfig struct {
        Name    string	    `toml:"name"`
        Ssl     bool        `toml:"ssl"`
        Host    string      `toml:"host"`
        Hosts   []string    `toml:"hosts"`
        Weights []int       `toml:"weights"`
        
        Charset	string	`toml:"charset"`
        Domain	string	`toml:"domain"`
        Cookie	string	`toml:"cookie"`
        Expiry	string  `toml:"expiry"`
        MaxAge	string  `toml:"maxage"`
        Crypto	string  `toml:"crypto"`

        Setting Map     `toml:"setting"`
    }

    ViewConfig struct {
        Driver  string  `toml:"driver"`
        Left    string  `toml:"left"`
        Right   string  `toml:"right"`
        Setting Map     `toml:"setting"`
    }

    PathConfig struct {
        Node        string    `toml:"node"`
        Lang        string    `toml:"lang"`
        View        string    `toml:"view"`
        Static      string    `toml:"static"`
        Plugin      string    `toml:"plugin"`
        Upload      string    `toml:"upload"`
        Shared      string    `toml:"shared"`
        Storage     string    `toml:"storage"`
        Thumbnail   string    `toml:"thumbnail"`
    }

)

func (config *configConfig) Langs(extens ...Map) Map {
    langs := Map{}

    for k,v := range config.Lang {
        langs[k] = v.Name
    }

    if len(extens) > 0 {
        for k,v := range extens[0] {
            langs[k] = v
        }
    }

    return langs
}