package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type PingResult struct {
	Addr        string
	PacketsSent int
	PacketsRecv int
	PacketLoss  float64
	MinRtt      time.Duration
	AvgRtt      time.Duration
	MaxRtt      time.Duration
	StdDevRtt   time.Duration
}

func pingAddress(addr string, wg *sync.WaitGroup, results chan<- PingResult) {
	defer wg.Done()
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	// Configurar para usar modo privilegiado
	pinger.SetPrivileged(true)

	// Configuración del pinger
	pinger.Count = 100 // Enviar 100 paquetes
	pinger.Timeout = 15 * time.Second // Aumenta el tiempo de espera total
	pinger.Interval = 100 * time.Millisecond // Intervalo entre pings

	// Manejo del ping
	err = pinger.Run()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	stats := pinger.Statistics()
	result := PingResult{
		Addr:        addr,
		PacketsSent: stats.PacketsSent,
		PacketsRecv: stats.PacketsRecv,
		PacketLoss:  stats.PacketLoss,
		MinRtt:      stats.MinRtt,
		AvgRtt:      stats.AvgRtt,
		MaxRtt:      stats.MaxRtt,
		StdDevRtt:   stats.StdDevRtt,
	}
	results <- result
}

func main() {
	addresses := []string{"panel.tensolite.com", "www.tensolite.com", "proveedores.tensolite.com"} // Puedes ajustar las IPs o URLs aquí
	results := make(chan PingResult, len(addresses))
	var wg sync.WaitGroup

	// Record the start time
	start := time.Now()

	// Iniciar el ping
	for _, addr := range addresses {
		wg.Add(1)
		go pingAddress(addr, &wg, results)
	}

	// Cerrar el canal de resultados cuando todos los pings hayan terminado
	go func() {
		wg.Wait()
		close(results)
	}()

	// Manejo de la señal Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		fmt.Println("\nInterrupt signal received, exiting...")
		os.Exit(0)
	}()

	// Imprimir resultados
	for result := range results {
		fmt.Printf("\n--- Ping statistics for %s ---\n", result.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %.2f%% packet loss\n",
			result.PacketsSent, result.PacketsRecv, result.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			result.MinRtt, result.AvgRtt, result.MaxRtt, result.StdDevRtt)
	}

	// Imprimir el tiempo total
	totalDuration := time.Since(start)
	fmt.Printf("\nTotal time for all pings: %v\n", totalDuration)

	// Esperar a que el usuario presione Enter para salir
	fmt.Println("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
