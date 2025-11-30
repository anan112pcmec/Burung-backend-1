package config

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Environment struct {
	DBHOST, DBUSER, DBPASS, DBNAME, DBPORT           string
	RDSHOST, RDSPORT                                 string
	RDSENTITYDB, RDSBARANGDB, RDSENGAGEMENTDB        int
	MEILIHOST, MEILIKEY, MEILIPORT                   string
	RMQ_HOST, RMQ_USER, RMQ_PASS, EXCHANGE, RMQ_PORT string

	CASSANDRA_HOST     string
	CASSANDRA_PORT     string
	CASSANDRA_KEYSPACE string
}

func (e *Environment) RunConnectionEnvironment() (
	db *gorm.DB,
	redis_entity *redis.Client,
	redis_barang *redis.Client,
	redis_engagement *redis.Client,
	search_engine meilisearch.ServiceManager,
	notification *amqp091.Connection,
	cassx gocqlx.Session,
) {

	/*
		POSTGRESQL
	*/
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		e.DBHOST, e.DBUSER, e.DBPASS, e.DBNAME, e.DBPORT,
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("❌ PostgreSQL error: %v", err)
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ PostgreSQL ping: %v", err)
	}

	/*
		REDIS
	*/
	redis_entity = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", e.RDSHOST, e.RDSPORT),
		DB:   e.RDSENTITYDB,
	})

	redis_barang = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", e.RDSHOST, e.RDSPORT),
		DB:   e.RDSBARANGDB,
	})
	redis_engagement = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", e.RDSHOST, e.RDSPORT),
		DB:   e.RDSENGAGEMENTDB,
	})

	/*
		RABBITMQ
	*/
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", e.RMQ_USER, e.RMQ_PASS, e.RMQ_HOST, e.RMQ_PORT)
	notification, err = amqp091.Dial(connStr)
	if err != nil {
		log.Fatalf("❌ RabbitMQ error: %v", err)
	}

	/*
		MEILISEARCH
	*/
	search_engine = meilisearch.New(
		fmt.Sprintf("http://%s:%s", e.MEILIHOST, e.MEILIPORT),
		meilisearch.WithAPIKey(e.MEILIKEY),
	)

	/*
		CASSANDRA
	*/
	log.Println("🟣 Connecting to Cassandra...")

	cluster := gocql.NewCluster(fmt.Sprintf("%s:%s", e.CASSANDRA_HOST, e.CASSANDRA_PORT))
	cluster.Keyspace = e.CASSANDRA_KEYSPACE
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	cassx, err = gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatalf("❌ gocqlx WrapSession: %v", err)
	}

	return
}
