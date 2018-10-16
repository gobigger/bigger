


global注册  *.index
sys 注册 sys.index

有时候访问sys.index，还是显示全局的
好像注册的顺序有问题，sys.index没有覆盖*.index中的sys.index



view层template支持定义
Bigger.Template(name, default)
这样就可以在不使用vfs的情况下， 内置一些文档的VIEW


service的定义，直接直接event, queue 的调用，如

Bigger.Register("name", Map{
    "event": true, "queue": true,
    "name": name, "text": "text",
    "action": action,
})



sv.Invoke的时候，连带setting



<!-- auth, item 的empty,error处理，auth完成,item的功能好像已经去掉不使用了 -->


<!-- http.setcookie的时候，域名自动使用当前域名的根域名，以保证在额外使用域名的时候，cookie有效 -->


<!-- mapping的时候， 添加上下文， 以及支持返回客户端时间的时候，自定义时区 -->


<!-- built-in
系统自身是要什么types来着，一定要built-in？？？
ctx.Answer，那就不需要内置了 types, 这些了，放到builtin包里去 
不过state, lang这些还是有系统的
-->




<!-- 为了方便动态加载，各模块在注册Router的时候， config不要带进去，只注册基本的参数，比如，名称，时间，URI，域名什么的
这样请求过来的时候， 从系统里拉config， 就可以保证config是最新的，而不用重新注册 -->

<!-- 当然，如果有新动态加载，那就需要重新注册， 如果要记录一下， 已经注册的对象。
这样在再次调用Register时（动态加载），要先解除原有的注册，然后重新注册
系统只调用Register注册，判断已经注册需要在驱动里实现。 -->



<!-- 注册，HTTP分method的时候， args, auth节要合并，而不是覆盖 -->


<!-- 启动时加载plugins，已经可以了
但是一定要预加载， 要不然，如果数据模型写在so里，路由引用参数的时候，就引用不到了
因为代码会优先执行。 但是在bigger.init的时候，加载so会有问题，因为so也会加载bigger，就重复init会出错
使用另外一个加载包解决了此问题 -->


<!-- event,queue Start可以简化， 因为所有连接都有 生产和消费
所以，一直Start方法就可以了 -->


<!-- http模块的动态Register还没好。其它模块好了 -->


<!-- 各默认驱动不要放在主包， 以减小主包体积， 看是不是编译出来的SO文件会小一些。
不光驱动， 还有其它可能移到子包的代码，都移开。。 尽量减小主包体积 -->


<!-- Plan模块，time 改成 timer ，   而且要改成Timer可以单独注册， 
这样可以动态的只添加timer，而router不动。 因为要执行计划就几个， 但是可以动态设置什么时候执行 -->


<!-- 各模块驱动Register方法需要处理重新注册的情况
如果有新注册（已经存在老的，需要删除老的先），如果不能删除，考虑覆盖 -->


<!-- mapping支持多名称，如type,types -->


<!-- 触发器模块 -->


<!-- config.path 还有待考虑， 应该直接融合到各模块配置。不需要单独拿一个出来定义path
lang的目录定义貌似还没有找到合适的地方，直接干掉，没有自定义的意义 -->


<!-- HTTP单端口多域名化，这样可以简化部署，考虑一下，完成 -->






<!-- form中要请求的语言，按浏览器携带的语言匹配 -->



<!-- plan.Timer 注册到 Router中的一个字段。  而不单独用一个branch -->



<!-- data.Fields 相关的方法  -->



<!-- langs 中配置中的默认目录，现在是写死的langs， 要可以在配置中自定义
statics 的默认配置， plugins 目录的配置  shared -->




<!-- Router注册代码，不需要返回， 各模块注册都要去掉一下
生成文档走 Bigger 对象统一返回数据 -->


<!-- http.filter, http.handler 支持*分开注册。注册时候已经分开， -->


<!-- 几个模块的 routerActions 这些方法都是重复代码， 考虑封装成一个方法 完成 -->


<!-- 表单处理，空文件也被处理成了Map bug -->



<!-- serve 拦截器 考虑下存在的必要
request -> form -> args -> auth -> execute -> response ->  -->




<!-- 文件模块的 PublicUrl  ThumbnailUrl
获取URL方法已经实现， 还要考虑生成缩图的代码， 光有  Preview 的链接获取不够
文件模块要在base.close对所有 读取的对象要关闭， 所以ctx.final，要在body完了之后执行
Browse方法，要可以传自定义的文件名 -->

<!-- 存储文件安全访问，加过期时间等验证。。
Browse要自带name参数吗？ -->



<!-- bigger是否要带 hashid  hash64 ，应该要统一一下，有好些地方可以用， 基本上可以做为简单的加密校验了 -->




<!-- 文件token还有问题 -->



<!-- view中的 browse preview 方法 -->


<!-- filebase多点配置，加入权重weight，在 FILE.Assign 不指定库的时候，自动按权重分配一个存储池
这样就可以自动分散到不同的目录里去， 比如，多台服务器节点共享网络驱动器。
注意， weight=0的存储库，不参与随机分散。 -->




<!-- File.Assign 考虑是否带metadata，有部分存储系统应该支持这个 -->




<!-- FILE模块的水印功能，要不要支持文字，因为支持文字就要字段文件，自行写文件下载方法，打上水印，或是弄成缩图？ -->
<!-- 还有像音频/视频的压缩不同的质量，打水印什么的。这个不做为文件模块的功能，应该放到业务层处理 -->


<!-- websocket支持，框架第3版左右的时候，写过websocket的模块
nats做为消息中心，如同event模块，所有连接进来，就直接订阅nats对应的消息
要记录连接者的id，做为单个消息订阅
还可以订阅分组的消息。
websocket分2种定义，  一个是消息Message，一个是命令Command
消息是指服务器下发给客户端的，Message表示，会发给客户端什么样的消息
命令是指客户端发给服务端的，因为Command表示，服务端支持哪些命令
http.ctx.Accept 表示接受连接？
处理器还有 connect, disconnect, 收听， 取消。 出错，是不是放在触发器里？



方案一， 连接只到进程，收到消息或广播，由进程处理再发给客户端，
分多个pub/sub服务器， 每个进程都连接所有服务器， 
订阅一个默认频道，然后订阅一个自己进程NAME的频道，
收到来自客户端的消息，直接是在进程内完成的。 不用经过socket
只有服务器发给客户端的消息，才需要经过消息服务器。因为不知道客户端连接在哪台服务器 -->


<!-- Upgrade如果id已经存在， 就考虑踢掉老的，或是不上新的 -->


<!-- 频道的订阅退订还没处理 -->


<!-- websocket可能不需要指定bases，因为id按自动分片来比较好。
要不然，订阅channel的时候， 也得指定， 那就得想办法把id加密，像filecode一样
可以考虑在Upgrade前加一个，类似Assgin方法，拿到一个ID（是原ID+base）编码后的加密串
然后在其它方法里，就可以用这个串来解析，是属于哪个base -->




<!-- 方案二，客户端连接时，直接连接消息服务器，订阅自己ID和对应频道的广播。
这样消息服务器应该无法承受。比如同时100万在线， 消息服务器应该会疯的
而且要扩容也不好弄。。。。  抛弃此方案。。 -->



<!-- firefox COOKIE貌似无效，每次请求都是新Id -->


<!-- cookie读写自动加解密 -->


<!-- 一个根Context， 具体的用具体的Context，比如
HttpContext,  QueueContext
服务层本来就不使用Context了。其它地方还是要使用统用的Context -->


<!-- 各种xxxBrach都没有用了，直接用一个coreBrach带name就完了 -->


<!-- raft分等级日志。 -->


<!-- ctx.Down 
ctx.Buffer -->

<!-- 所有驱动prefix自动设为 name -->

<!-- 各Branch都没用， 直接用一个coreBranch， N多实例就可以完成了 -->

<!-- plan在raft选举期间就无法执行，需要考虑在选举期间先缓存执行列表
等到选举完成时，把列表里的计划再执行一遍。已经使用Delay延期执行
并且在计划定义的时候，可用参数 "delay": false，关闭某计划的延期执行
比如，每秒拉行情的是不允许延期执行 -->

<!-- 404拿不到当前site -->

<!-- file.Browse加参数 -->

<!-- 写一个可以按文件大小，日期，行数来分日志文件的日志驱动 -->

<!-- 日志加一个配置，用来做为标识，主要是用来标记是哪个节点，记录Bigger.Id什么的。 -->

<!-- 队列， 事件，  本身也可以用hashring来处理，具体要发给哪个库 -->

<!-- event/trigger合并为event
trigger只在进程内触发 publish广播给所有进程 -->

<!-- 内存版的session和cache 要走一个内存版的k/v库去处理 -->

<!-- 内存版的K/V， 用于 session和cache
文件版的K/V， 用于 session和cache -->

<!-- redis会话驱动 -->

<!-- 所有注册可不覆盖，override = false，bigger/builtin里的部分已完成，全部完成 -->

<!-- ctx.Error, ctx.Failed, ctx.Denied 参数更新为  *Error -->

<!-- 直接定义状态错误,StateError(code, state, text)自动生成状态码和String -->

<!-- ctx加一个执行结果，有以下几种
found, error, failed, denied, succeed
这样在输出logger的时候，可以知道哪一次请求的最终执行结果 -->

<!-- 所有驱动*Error考虑改回来error -->

<!-- redis缓存驱动 -->

<!-- redis事件，基本完成，动态加载也可以了。 -->

<!-- redis队列驱动，主要代码已经完成，  还差热加载。
驱动感觉有问题，CPU很高， 已经解决， 热加载也解决了。 -->

<!-- Error对象，直接改成 Status，然后所有对象直接使用error本身。
但是ctx.lastError需要记录，要有状态和参数信息
ctx.Erred() 还是用 error 不变，  多加一个  Status字段？
Bigger.Mapping这方法返回的是 *Error， 这里好像必须要一个Error对象
因为如果返回error，参数就不好带回来，不知道是哪个参数出错了 -->

<!-- config中所有模块的默认前缀，不能有， 不能按节点名来指定，这样在多节点就没法通用了 -->

<!-- 会话模块，可以像file一样，支持多个连接
然后在框架层使用hashring来按id来决定使用哪个连接
方便在大规模时分散压力，这样的话需要所有的节点都同一配置
比如，redis不搞集群的时候，可以这样软件分散，或是直接在驱动里实现吧 -->



<!-- Fatal 一般是输出错误，然后退出程序。。  logger不能这么用，咱不需要退出 -->



////语言包加载的顺序是在init里， 所以代码和so里的新东西（主要是default默认的，会无效加载，会被代码替换的可能）



生成系统文档



#postgresql://cockroachdb01.vpc.aws:26257,cockroachdb02.vpc.aws:26257,cockroachdb03.vpc.aws:26257/test?loadBalanceHosts=true




memcache的缓存驱动 Keys 尚未实现



验证码功能



虚拟文件系统，Statics, Views 都可以放入内存，再配合文件监控，动态加载文件



<!-- view层直接string做为模板。 OK -->


<!-- Logger驱动的 level处理 -->




<!-- mutex 模块，提供内存锁，或是分布式锁 -->
<!-- mutex  redis驱动
mutex  memcache驱动 -->





<!-- filebase, database, cachebase
.Base的时候，默认可为空，为空时，如果只有一个配置，就直接拿那一个，简化开发
如果有多个的时候，使用默认的配置的 -->



<!-- logic层的请求封装，封好ctx,args,setting等对象， 每一次调用都一个req
req依赖Context对象. 还可以考虑异步返回， 
将来可以直接升级logic为服务层，直接接入HTTP，把所有服务暴露出去。或是按设置暴露


逻辑层考虑改个名称？或是不改。 
三方调用的方式，也都可以并入逻辑层，统一方法调用。？
各种可以软驱动的都可以直接放到logic模块注册。
比如：逻辑层，三方调用，支付，  区块链调用（可考虑独立模块是否有必要）


pay := ctx.Service("pay.unionpay", setting)
result := pay.Invoke("charge", args)


比如， 
短信发送，  可能有N多家供应商
邮件发送，  多家供应商
呼叫验证码  多家
支付通道    多家
区块钱包    多种
 -->





<!-- 模块全小写包里可访问，统一入口Bigger.XXXX
大概完成， 还有些命名可能要考虑一下。比如， Encoding/Decoding/ -->




<!-- view层的 backurl 什么的，系统自带函数库 -->





路由配置中考虑加入session配置节，为session数据mapping一下。
或者放到args里？value里？
因为大多k/v都是基于string的，所以有JSON解析后，类型与写入的不一样
所以来一个SESSION的定义，好让SESSION解析后方便在代码里使用
或是把session也接入args







定时任务，现在直接在代码里 time.Sleep，这样如果程序退了，或是重启了， 这一部分的就没有执行
或者可以考虑来一个模块， 把定义信息都缓存起来，到点自动清理， 重启的时候，拉回来来跑一下？ 那样func就得写死，不能用局部func了



可考虑把statics, views, 加载到内存虚拟文件系统， 然后监听目录的文件修改操作， 自动更新内存中的文件
这样可以极大的提高views, statics 的性能， 内存占用也大不了多少。
动态加载插件， 监听plugins目录的文件



url.route中的args处理的时候，要简化，
要不然，加密一个参数，得写 encode=xx, decode=xx 写2
还有可能的性能问题，参数太多的时候， 只根据URL中的参数来处理一下，要简化的多
query参数也要考虑如果ARGS中有，加解密的处理



还可以优化的地方，filter,handler方法的获取，可以加“缓存”，不要每次都从core里获取定义
这样可以稍稍的再少点代码调用  --这个还要考虑动态加载，不过也很好实现
还有其它内部功能的缓存化，获取数据时候的缓存化，可以省一点代码执行，意义不太大，但是可以极致优化



<!-- 大杀器
是否可以把框架本身的所有方法，都注册进核心，这样构架就相当于只有骨架。
随时可以用新的方法去重新定义框架内置的方法代码，不过估计不行，因为好多内部变量，外部无法访问 -->



fund数据库设计要划分一下
    fund_ 只用来记录 帐户，交易，流水，
    区块链这部分， 有充值，提现，也要用到帐户体系
    支付通道，也有充值，提现， 也要用到帐户体系
    不过同一套系统， 既使用法币，又使用数字币的不多



风控功能历史化，
比如，买入功能，限制时间，触发条件。为一条记录。   
一条记录限制用户的某一项功能， 系统设置一些触发条件，自动或手动触发风控。
必考虑支持分组， 为某一类用户风控。 比如，商户被风控，旗下的客户就被限制。
这样所有的风控都有历史记录，谁什么时候被怎么样。




缓存也可以按hashring来先在框架层面来分散一下
还是放到驱动里自行分散的好？ 考虑到redis不搞集群，而是搞一堆手动分散
应该放到驱动里分散，比较好一点。




有执行线的模块,event,queue,plan,http，考虑使用同一个父struct，
这样可以共用很多代码， http 可以单独重写对应的代码， 这样可以简化代码量
event,queue,plan，的执行线代码是完全一样的，可以合成一个