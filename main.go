package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// держим веб-сервер жить, пока не прервут процесс
func waitIfServeKeep(srv *httpServer) {
	if serveKeep {
		fmt.Println("web server staying up. Press Ctrl+C to stop.")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	}
	_ = srv.Shutdown()
}

func main() {
	// флаги работы чекера
	flag.IntVar(&workers, "workers", runtime.NumCPU()*2, "number of parallel workers")
	flag.DurationVar(&bootWait, "boot-wait", DefaultBootWait, "wait after xray start")
	flag.DurationVar(&testTimeout, "test-timeout", DefaultTestTimeout, "HTTP test timeout")
	flag.DurationVar(&xrayRunBudget, "xray-budget", DefaultXrayRunBudget, "per-check time budget")
	flag.IntVar(&retrySNI, "retry-sni", 3, "max SNI attempts per config")
	flag.BoolVar(&enableTCPProbe, "tcp-probe", true, "fast TCP probe before starting xray")
	flag.IntVar(&maxWorkCfg, "maxworkcfg", 0, "stop after N working configs (0 = unlimited)")
	flag.BoolVar(&serveKeep, "serve-keep", false, "keep web server running after checks finish")
	flag.Parse()

	// конфиг веба + старт сервера
	cfg, err := LoadAppConfig("config.json")
	if err != nil {
		// не фатально: создадим шаблон и продолжим
		fmt.Println("web config:", err)
	}
	srv := StartWebServer(cfg)

	// 3. САМОЕ ГЛАВНОЕ: Просто запускаем скан напрямую. 
	// Никаких 'go', никаких фонов. Программа будет работать, пока не закончит.
	fmt.Println("--- STARTING SCAN ---")
	RunScanOnce(maxWorkCfg) 
	fmt.Println("--- SCAN FINISHED, SAVING RESULTS ---")

	// 4. Принудительно закрываем сервер и выходим
	_ = srv.Shutdown()
	_ = os.Stdout.Sync()
	
	// Программа завершится сама здесь
}
