package main

import (
	
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
	"net"
)

type Scenario string

const (
	ScenarioNormal      Scenario = "normal"
	ScenarioBeforeStart Scenario = "before-start"
	ScenarioHalfway     Scenario = "halfway"
	ScenarioMostlyDone  Scenario = "mostly-done"
	ScenarioIdle        Scenario = "idle"
)

const (
	halfwayMessage    = messageCount / 2
	mostlyDoneMessage = messageCount * 9 / 10

	shutdownTestTimeout = 30 * time.Second
	serverStartTimeout  = 5 * time.Second
)

type ScenarioResult struct {
	Scenario       Scenario
	Duration       time.Duration
	ServerShutdown error
	ClientResults   []Result
}

func runScenario(scenario Scenario) (ScenarioResult, error) {
	startTime := time.Now()

	server, err := startServer()
	if err != nil {
		return ScenarioResult{}, err
	}

	defer func() {
		if server.ProcessState == nil {
			_ = server.Process.Kill()
		}
	}()

	switch scenario {
	case ScenarioNormal:
		results, err := runNormalScenario()
		if err != nil {
			return ScenarioResult{}, err
		}

		return ScenarioResult{
			Scenario:     scenario,
			Duration:     time.Since(startTime),
			ClientResults: results,
		}, nil

	case ScenarioBeforeStart:
		return runBeforeStartScenario(server, startTime)

	case ScenarioHalfway:
		return runProgressScenario(
			server,
			startTime,
			scenario,
			halfwayMessage,
		)

	case ScenarioMostlyDone:
		return runProgressScenario(
			server,
			startTime,
			scenario,
			mostlyDoneMessage,
		)

	case ScenarioIdle:
		return runIdleScenario(server, startTime)

	default:
		return ScenarioResult{}, fmt.Errorf(
			"unknown scenario: %s",
			scenario,
		)
	}
}

func startServer() (*exec.Cmd, error) {
	cmd := exec.Command(
		"go",
		"run",
		"./cmd/server",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	// Wait until the server is accepting connections.
	deadline := time.Now().Add(serverStartTimeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout(
			"tcp",
			serverAddress,
			100*time.Millisecond,
		)

		if err == nil {
			conn.Close()
			return cmd, nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	_ = cmd.Process.Kill()

	return nil, fmt.Errorf("server did not start within %s", serverStartTimeout)
}

func runNormalScenario() ([]Result, error) {
	var wg sync.WaitGroup

	results := make(chan []Result, clientCount)

	eventCh := make(chan ClientEvent, clientCount*messageCount)

	for clientID := 1; clientID <= clientCount; clientID++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			results <- runClient(
				serverAddress,
				id,
				eventCh,
				nil,
				0,
			)
		}(clientID)
	}

	wg.Wait()

	close(results)
	close(eventCh)

	var allResults []Result

	for clientResults := range results {
		allResults = append(allResults, clientResults...)
	}

	return allResults, nil
}
func runBeforeStartScenario(
	server *exec.Cmd,
	startTime time.Time,
) (ScenarioResult, error) {

	eventCh := make(chan ClientEvent, clientCount)

	results := make(chan []Result, clientCount)

	var wg sync.WaitGroup

	for clientID := 1; clientID <= clientCount; clientID++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			results <- runClient(
				serverAddress,
				id,
				eventCh,
				nil,
				0,
			)
		}(clientID)
	}

	connected := 0

	for connected < clientCount {
		event := <-eventCh

		if event.Type == ClientConnected {
			connected++
		}
	}

	fmt.Printf(
		"All %d clients connected. Triggering shutdown...\n",
		clientCount,
	)

	shutdownErr := shutdownServer(server)

	wg.Wait()

	close(results)

	var allResults []Result

	for clientResults := range results {
		allResults = append(allResults, clientResults...)
	}

	return ScenarioResult{
		Scenario:       ScenarioBeforeStart,
		Duration:       time.Since(startTime),
		ServerShutdown: shutdownErr,
		ClientResults:  allResults,
	}, nil
}

func runProgressScenario(
	server *exec.Cmd,
	startTime time.Time,
	scenario Scenario,
	threshold int,
) (ScenarioResult, error) {

	eventCh := make(chan ClientEvent, clientCount*messageCount)

	results := make(chan []Result, clientCount)

	var wg sync.WaitGroup

	progress := make([]int, clientCount+1)

	for clientID := 1; clientID <= clientCount; clientID++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			results <- runClient(
				serverAddress,
				id,
				eventCh,
				nil,
				0,
			)
		}(clientID)
	}

	fmt.Printf(
		"Waiting for every client to reach message %d...\n",
		threshold,
	)

	for {
		event := <-eventCh

		switch event.Type {

		case MessageCompleted:

			if event.MessageID > progress[event.ClientID] {
				progress[event.ClientID] = event.MessageID
			}

			if everyClientReached(progress, threshold) {
				goto shutdown
			}
		}
	}

shutdown:

	fmt.Printf(
		"Every client reached message %d. Triggering shutdown...\n",
		threshold,
	)

	shutdownErr := shutdownServer(server)

	wg.Wait()

	close(results)

	var allResults []Result

	for clientResults := range results {
		allResults = append(allResults, clientResults...)
	}

	return ScenarioResult{
		Scenario:       scenario,
		Duration:       time.Since(startTime),
		ServerShutdown: shutdownErr,
		ClientResults:  allResults,
	}, nil
}

func everyClientReached(
	progress []int,
	threshold int,
) bool {

	for clientID := 1; clientID <= clientCount; clientID++ {
		if progress[clientID] < threshold {
			return false
		}
	}

	return true
}

func runIdleScenario(
	server *exec.Cmd,
	startTime time.Time,
) (ScenarioResult, error) {

	connections := make([]net.Conn, 0, clientCount)

	fmt.Printf("Opening %d idle connections...\n", clientCount)

	for clientID := 1; clientID <= clientCount; clientID++ {

		conn, err := net.DialTimeout(
			"tcp",
			serverAddress,
			clientTimeout,
		)

		if err != nil {
			fmt.Printf(
				"client %d failed to connect: %v\n",
				clientID,
				err,
			)
			continue
		}

		connections = append(connections, conn)
	}

	fmt.Printf(
		"Idle connections established: %d/%d\n",
		len(connections),
		clientCount,
	)

	// Give the server time to put all these
	// connections into Read().
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Triggering shutdown with idle clients...")

	shutdownErr := shutdownServer(server)

	for _, conn := range connections {
		_ = conn.Close()
	}

	return ScenarioResult{
		Scenario:       ScenarioIdle,
		Duration:       time.Since(startTime),
		ServerShutdown: shutdownErr,
	}, nil
}
func shutdownServer(cmd *exec.Cmd) error {

	if cmd.Process == nil {
		return fmt.Errorf("server process is not running")
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("send interrupt: %w", err)
	}

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {

	case err := <-done:
		return err

	case <-time.After(shutdownTestTimeout):
		_ = cmd.Process.Kill()

		return fmt.Errorf(
			"server did not shut down within %s",
			shutdownTestTimeout,
		)
	}
}