package main

import (
	"expvar"
	"fmt"
	"runtime"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/yanpavel/social_project/internal/auth"
	"github.com/yanpavel/social_project/internal/db"
	"github.com/yanpavel/social_project/internal/env"
	"github.com/yanpavel/social_project/internal/mailer"
	"github.com/yanpavel/social_project/internal/store"
	"github.com/yanpavel/social_project/internal/store/cache"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Social Network API
//	@description	API for social network
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendUrl: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:6404/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:          env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:            env.GetString("REDIS_PW", ""),
			db:            env.GetInt("REDIS_DB", 0),
			enabled:       env.GetBool("REDIS_ENABLED", true),
			isRatelimiter: env.GetBool("RATELIMITER_ENABLED", true),
		},
		env: env.GetString("ENV", "production"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "demo@demomailtrap.co"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_APIKEY", "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"),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_APIKEY", "d308b01b30ca9c36e9d428e8c7fc08fd"),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "no_moresecrets"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "lesocial",
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		fmt.Printf("%q", err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	// Authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss,
	)

	// Cache Redis
	var rdb *redis.Client
	var limiter *redis_rate.Limiter
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis connection established")

		// Limiter Redis
		if cfg.redisCfg.isRatelimiter {
			limiter = redis_rate.NewLimiter(rdb)
		}
	}

	store := store.NewStorage(db)

	cacheStore := cache.NewRedisStorage(rdb)

	mailer, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	app := &application{
		config:           cfg,
		store:            store,
		cacheStorage:     cacheStore,
		logger:           logger,
		mailer:           mailer,
		authenticator:    jwtAuthenticator,
		redisRateLimiter: *limiter,
	}

	expvar.NewString("version")
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
