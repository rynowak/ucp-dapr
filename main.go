package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	daprclient "github.com/dapr/go-sdk/client"
	daprservice "github.com/dapr/go-sdk/service/http"
	daprworkflow "github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/api"
	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/reconciler"
	"github.com/rynowak/ucp-dapr/pkg/resources"
	"github.com/rynowak/ucp-dapr/pkg/rp/containers"
	"github.com/rynowak/ucp-dapr/pkg/subscribe"
)

func main() {
	fmt.Printf("Connecting to Dapr\n")
	dapr, err := daprclient.NewClient()
	if err != nil {
		log.Fatalf("error creating Dapr client: %v", err)
	}

	defer dapr.Close()

	db.Client = dapr
	subscribe.Client = dapr

	worker, err := registerWorkflows(dapr)
	if err != nil {
		log.Fatalf("error registering Dapr workflow: %v", err)
	}

	err = worker.Start()
	if err != nil {
		log.Fatalf("error starting Dapr workflow worker: %v", err)
	}

	server := createServer(dapr)

	service := daprservice.NewService(":8081")
	for _, subscription := range subscribe.Subscriptions {
		copy := subscription
		service.AddTopicEventHandler(&copy, subscribe.ResourceEvent)
	}

	go func() {
		err := service.Start()
		if err != nil {
			log.Fatalf("error starting Dapr pubsub listener worker: %v", err)
		}
	}()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		service.GracefulStop()
		server.Shutdown(context.Background())
		worker.Shutdown()
	}()

	go func() {
		test(ctx)
	}()

	fmt.Printf("Server is running on %s\n", server.Addr)
	err = server.ListenAndServe()
	if errors.Is(http.ErrServerClosed, err) {
		log.Println("server shutdown gracefully")
		os.Exit(0)
	} else {
		log.Println("server error", err)
		os.Exit(1)
	}
}

func createServer(dapr daprclient.Client) *http.Server {
	handler := &api.Handler{
		Dapr:           dapr,
		StateStoreName: "statestore",
		PubSubName:     "pubsub",
		ResourceType:   "Applications.Core/containers",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /planes/radius/{planeName}/resourceGroups/{resourceGroupName}/providers/Applications.Core/containers", handler.ListHandler)
	mux.HandleFunc("GET /planes/radius/{planeName}/resourceGroups/{resourceGroupName}/providers/Applications.Core/containers/{name}", handler.GetHandler)
	mux.HandleFunc("DELETE /planes/radius/{planeName}/resourceGroups/{resourceGroupName}/providers/Applications.Core/containers/{name}", handler.DeleteHandler)
	mux.HandleFunc("PUT /planes/radius/{planeName}/resourceGroups/{resourceGroupName}/providers/Applications.Core/containers/{name}", handler.PutHandler)

	mux.HandleFunc("GET /planes/radius/{planeName}/providers/{namespace}/operationStatuses", handler.OperationStatusListHandler)
	mux.HandleFunc("GET /planes/radius/{planeName}/providers/{namespace}/operationStatuses/{name}", handler.OperationStatusGetHandler)

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}

func registerWorkflows(dapr daprclient.Client) (*daprworkflow.WorkflowWorker, error) {
	worker, err := daprworkflow.NewWorker(daprworkflow.WorkerWithDaprClient(dapr))
	if err != nil {
		return nil, fmt.Errorf("error creating Dapr workflow worker: %w", err)
	}

	err = worker.RegisterWorkflow(reconciler.Reconcile)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr workflow: %w", err)
	}

	err = worker.RegisterWorkflow(containers.ContainerPut)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr workflow: %w", err)
	}

	err = worker.RegisterWorkflow(containers.ContainerPut)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr workflow: %w", err)
	}

	err = worker.RegisterActivity(reconciler.CheckResourceExistance)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr activity: %w", err)
	}

	err = worker.RegisterActivity(reconciler.FetchCurrentGeneration)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr activity: %w", err)
	}

	err = worker.RegisterActivity(reconciler.SaveResource)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr activity: %w", err)
	}
	err = worker.RegisterActivity(reconciler.CommitOperation)
	if err != nil {
		return nil, fmt.Errorf("error registering Dapr activity: %w", err)
	}

	return worker, nil
}

func test(ctx context.Context) {
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * 3)

		fmt.Println("=====Starting=====")
		fmt.Println()

		log.Println("=====Creating containers=====")
		err := createContainer(ctx, "A", map[string]any{})
		if err != nil {
			log.Println("error creating container", err)
			continue
		}

		err = createContainer(ctx, "B", map[string]any{})
		if err != nil {
			log.Println("error creating container", err)
			continue
		}
		log.Println("==========")
		log.Println()

		log.Println("=====Getting containers=====")
		err = getContainer(ctx, "A")
		if err != nil {
			log.Println("error getting container", err)
			continue
		}

		err = getContainer(ctx, "B")
		if err != nil {
			log.Println("error getting container", err)
			continue
		}

		log.Println("==========")
		log.Println()

		log.Println("=====Listing containers=====")
		err = listContainers(ctx)
		if err != nil {
			log.Println("error listing containers", err)
			continue
		}

		log.Println("==========")
		log.Println()

		log.Println("=====Deleting containers=====")
		err = deleteContainer(ctx, "A")
		if err != nil {
			log.Println("error deleting container", err)
			continue
		}

		err = deleteContainer(ctx, "B")
		if err != nil {
			log.Println("error deleting container", err)
			continue
		}

		err = listContainers(ctx)
		if err != nil {
			log.Println("error listing containers", err)
			continue
		}

		log.Println("==========")
		log.Println()
	}
}

func createContainer(ctx context.Context, name string, properties any) error {
	body := map[string]any{
		"properties": properties,
	}
	bb, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", "http://localhost:8080/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name, bytes.NewReader(bb))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	bodyb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to create container: %s\n\n%s", resp.Status, string(bodyb))
	}

	log.Default().Println("response status", resp.Status)
	log.Default().Println("response body", string(bodyb))
	log.Default().Println("operation url", resp.Header.Get("Location"))

	err = pollOperation(ctx, "http://localhost:8080"+resp.Header.Get("Location"))
	if err != nil {
		return fmt.Errorf("error polling operation: %w", err)
	}

	return do(req)
}

func getContainer(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	return do(req)
}

func listContainers(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	return do(req)
}

func deleteContainer(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", "http://localhost:8080/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	bodyb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to delete container: %s\n\n%s", resp.Status, string(bodyb))
	}

	log.Default().Println("response status", resp.Status)
	log.Default().Println("response body", string(bodyb))
	log.Default().Println("operation url", resp.Header.Get("Location"))

	err = pollOperation(ctx, "http://localhost:8080"+resp.Header.Get("Location"))
	if err != nil {
		return fmt.Errorf("error polling operation: %w", err)
	}

	return nil
}

func pollOperation(ctx context.Context, url string) error {
	for {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("response error: %w", err)
		}

		defer resp.Body.Close()

		bodyb, _ := io.ReadAll(resp.Body)

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("failed to check operation status: %s\n\n%s", resp.Status, string(bodyb))
		}

		data := resources.OperationStatusResource{}
		err = json.Unmarshal(bodyb, &data)
		if err != nil {
			return fmt.Errorf("error unmarshalling response: %w", err)
		}

		log.Default().Println("response status", resp.Status)
		if data.Status == "Succeeded" || data.Status == "Failed" || data.Status == "Canceled" {
			return nil
		}

		time.Sleep(time.Second * 3)
	}
}

func do(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	bodyb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to create container: %s\n\n%s", resp.Status, string(bodyb))
	}

	log.Default().Println("response body", string(bodyb))

	return nil
}
