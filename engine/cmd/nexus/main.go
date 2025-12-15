package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nexus-engine/internal/core"
	"nexus-engine/internal/modules/kv"
	"nexus-engine/internal/modules/pubsub"
	"nexus-engine/internal/pkg/logger"
)

func main() {
	// 1. Реестр модулей
	// Чтобы добавить Queue, просто допишем: queue.NewModule()
	enabledModules := []core.Module{
		kv.NewModule(),
		pubsub.NewModule(),
	}

	// 2. Настройка флагов
	// Глобальные флаги
	port := flag.String("port", "4000", "Server port")
	logLevel := flag.Int("log-level", 1, "Log level (0=Error, 1=Info, 2=Debug)")

	// Регистрируем флаги каждого модуля
	for _, mod := range enabledModules {
		mod.RegisterFlags(flag.CommandLine)
	}

	flag.Parse()

	// 3. Инициализация
	log := logger.New(*logLevel)
	log.Info("🚀 Nexus Engine starting...")

	for _, mod := range enabledModules {
		log.Debug("Initializing module: %s", mod.Name())
		if err := mod.Init(log); err != nil {
			log.Error("Failed to init module %s: %v", mod.Name(), err)
			os.Exit(1)
		}
	}

	// 4. Роутинг
	mux := http.NewServeMux()

	// Health check (общий)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	// Регистрируем роуты модулей
	for _, mod := range enabledModules {
		mod.RegisterRoutes(mux)
	}

	// Middleware (Логирование)
	loggedMux := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Debug("HTTP %s %s took %v", r.Method, r.URL.Path, time.Since(start))
		})
	}(mux)

	// 5. Запуск сервера
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: loggedMux,
	}

	// Graceful Shutdown в отдельной горутине
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n🛑 Shutting down Nexus Engine...")

		// Останавливаем модули
		for _, mod := range enabledModules {
			log.Info("Stopping module: %s", mod.Name())
			mod.Shutdown()
		}

		os.Exit(0)
	}()

	log.Info("Ready on port %s", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Error("Server failed: %v", err)
	}
}
