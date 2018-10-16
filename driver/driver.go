package driver

import (
	_ "github.com/yatlabs/bigger/driver/logger-default"
	_ "github.com/yatlabs/bigger/driver/logger-file"
	_ "github.com/yatlabs/bigger/driver/mutex-default"
	_ "github.com/yatlabs/bigger/driver/mutex-memcache"
	_ "github.com/yatlabs/bigger/driver/mutex-redis"
	_ "github.com/yatlabs/bigger/driver/session-default"
	_ "github.com/yatlabs/bigger/driver/session-redis"
	_ "github.com/yatlabs/bigger/driver/session-memcache"
	_ "github.com/yatlabs/bigger/driver/session-file"
	_ "github.com/yatlabs/bigger/driver/session-memory"

	_ "github.com/yatlabs/bigger/driver/cache-default"
	_ "github.com/yatlabs/bigger/driver/cache-file"
	_ "github.com/yatlabs/bigger/driver/cache-memory"
	_ "github.com/yatlabs/bigger/driver/cache-memcache"
	_ "github.com/yatlabs/bigger/driver/cache-redis"
	_ "github.com/yatlabs/bigger/driver/data-postgres"
	_ "github.com/yatlabs/bigger/driver/data-cockroach"
	_ "github.com/yatlabs/bigger/driver/file-default"
	
	_ "github.com/yatlabs/bigger/driver/plan-default"
	_ "github.com/yatlabs/bigger/driver/event-default"
	_ "github.com/yatlabs/bigger/driver/event-redis"
	_ "github.com/yatlabs/bigger/driver/queue-default"
	_ "github.com/yatlabs/bigger/driver/queue-redis"

	_ "github.com/yatlabs/bigger/driver/http-default"
	_ "github.com/yatlabs/bigger/driver/view-default"
	_ "github.com/yatlabs/bigger/driver/socket-default"

)
