package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	routes "github.com/anan112pcmec/Burung-backend-1/app/Routes"
	maintain_cache "github.com/anan112pcmec/Burung-backend-1/app/cache/maintain"
	"github.com/anan112pcmec/Burung-backend-1/app/config"
	"github.com/anan112pcmec/Burung-backend-1/app/database/enums"
	"github.com/anan112pcmec/Burung-backend-1/app/database/migrate"
)

func Getenvi(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	rdsentity, _ := strconv.Atoi(Getenvi("RDSENTITY", "0"))
	rdsbarang, _ := strconv.Atoi(Getenvi("RDSBARANG", "0"))
	rdsengagement, _ := strconv.Atoi(Getenvi("RDSENGAGEMET", "0"))

	env := config.Environment{
		DBHOST:          Getenvi("DBHOST", "NIL"),
		DBUSER:          Getenvi("DBUSER", "NIL"),
		DBPASS:          Getenvi("DBPASS", "NIL"),
		DBNAME:          Getenvi("DBNAME", "NIL"),
		DBPORT:          Getenvi("DBPORT", "NIL"),
		RDSHOST:         Getenvi("RDSHOST", "NIL"),
		RDSPORT:         Getenvi("RDSPORT", "NIL"),
		RDSENTITYDB:     rdsentity,
		RDSBARANGDB:     rdsbarang,
		RDSENGAGEMENTDB: rdsengagement,
		MEILIHOST:       Getenvi("MEILIHOST", "NIL"),
		MEILIPORT:       Getenvi("MEILIPORT", "NIL"),
		MEILIKEY:        Getenvi("MEILIKEY", "NIL"),

		RMQ_HOST: "",
		RMQ_USER: "",
		RMQ_PASS: "",
		RMQ_PORT: "",
	}

	database, redis_entity_cache, redis_barang_cache, redis_engagement_cache, searchengine, _ :=
		env.RunConnectionEnvironment()

	// Router utama
	Router := mux.NewRouter()
	Router.Use(enableCORS)
	// Router.Use(rateLimitMiddleware)
	// Router.Use(blockBadRequestsMiddleware)

	// Jalankan enums dan migrasi
	if err := enums.UpEnumsEntity(database); err != nil {
		log.Printf("❌ Gagal UpEnumsEntity: %v", err)
	}
	if err := enums.UpBarangEnums(database); err != nil {
		log.Printf("❌ Gagal UpBarangEnums: %v", err)
	}
	if err := enums.UpEnumsTransaksi(database); err != nil {
		log.Printf("❌ Gagal UpEnumsTransaksi: %v", err)
	}

	migrate.UpEntity(database)
	migrate.UpBarang(database)
	migrate.UpTransaksi(database)
	migrate.UpEngagementEntity(database)
	migrate.UpSystemData(database)
	migrate.UpTresholdData(database)

	// Caching data
	maintain_cache.DataAlamatEkspedisiUp(database)
	maintain_cache.DataOperasionalPengirimanUp()
	//

	// Setup routes
	Router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	Router.PathPrefix("/").Handler(http.HandlerFunc(
		routes.GetHandler(database, redis_barang_cache, redis_entity_cache, searchengine),
	)).Methods("GET")

	Router.PathPrefix("/").Handler(http.HandlerFunc(
		routes.PostHandler(database, redis_entity_cache, redis_engagement_cache),
	)).Methods("POST")

	Router.PathPrefix("/").Handler(http.HandlerFunc(
		routes.PutHandler(database),
	)).Methods("PUT")

	Router.PathPrefix("/").Handler(http.HandlerFunc(
		routes.PatchHandler(database, redis_barang_cache, redis_engagement_cache),
	)).Methods("PATCH")

	Router.PathPrefix("/").Handler(http.HandlerFunc(
		routes.DeleteHandler(database),
	)).Methods("DELETE")

	// go cleanupClients()

	// Jalankan web server
	port := Getenvi("APPPORT", "8080")
	fmt.Printf("🚀 Server Burung berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(port, Router); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
