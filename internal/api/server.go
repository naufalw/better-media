package api

import (
	"better-media/internal/config"
	"better-media/internal/database"
	"better-media/internal/dispatcher"
	"better-media/internal/livestream"
	"better-media/internal/storage"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Config            *config.Config
	DB                *database.DB
	Storage           *storage.S3Client
	Dispatcher        dispatcher.JobDispatcher
	LivestreamManager *livestream.Manager
	Router            *gin.Engine
}

func NewServer(cfg *config.Config) (*Server, error) {
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	s3Client, err := storage.NewS3Client(cfg.S3BucketName, cfg.S3Endpoint, cfg.S3Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	var jobDispatcher dispatcher.JobDispatcher
	if cfg.DispatcherMode == "http" {
		if cfg.WorkerURL == "" || cfg.CallbackURL == "" {
			return nil, fmt.Errorf("WORKER_URL and CALLBACK_URL must be set when DISPATCHER_MODE=http")
		}
		jobDispatcher = dispatcher.NewHTTPDispatcher(s3Client, db, cfg.WorkerURL, cfg.CallbackURL)
		log.Println("Using external transcoder")
	} else {
		jobDispatcher = dispatcher.NewLocalDispatcher(s3Client, db)
		log.Println("Using local transcoder")
	}

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	server := &Server{
		Config:     cfg,
		DB:         db,
		Storage:    s3Client,
		Dispatcher: jobDispatcher,
		Router:     router,
	}

	lsManager := livestream.NewManager(cfg.RTMPPort, cfg.StreamsPath, s3Client, db, jobDispatcher)
	if err := lsManager.Start(); err != nil {
		log.Printf("Failed to start livestream manager: %v (livestream features disabled)", err)
	} else {
		server.LivestreamManager = lsManager
		log.Println("Livestream manager started")
	}

	server.registerRoutes()

	return server, nil
}

// Start HTTP
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.Config.Port)
	log.Printf("Starting server on %s", addr)
	return s.Router.Run(addr)
}

func (s *Server) Close() error {
	return s.DB.Close()
}
	return s.DB.Close()
}
