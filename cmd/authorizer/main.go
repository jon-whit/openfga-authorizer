package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/jon-whit/openfga-authorizer/internal/authorizer"
	webhookauthz "github.com/jon-whit/openfga-authorizer/internal/authorizer/webhook"
	"github.com/jon-whit/openfga-authorizer/internal/config"
	"github.com/jon-whit/openfga-authorizer/internal/controllers/clusterrolebindingsyncer"
	"github.com/jon-whit/openfga-authorizer/internal/controllers/clusterrolesyncer"
	"github.com/jon-whit/openfga-authorizer/internal/controllers/rolebindingsyncer"
	"github.com/jon-whit/openfga-authorizer/internal/controllers/rolesyncer"
	"github.com/jon-whit/openfga-authorizer/internal/nodeauth"
	"github.com/jon-whit/openfga-authorizer/internal/rbacconversion"
	"github.com/jon-whit/openfga-authorizer/internal/zanzibar"
	openfgasdk "github.com/openfga/go-sdk/client"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "the absolute file path to load the OpenFGA Authorizer config")

	flag.Parse()

	cfg := config.DefaultConfig()

	if configPath != "" {
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("failed to read config file: %v", err)
		}

		err = yaml.UnmarshalStrict(configBytes, &cfg)
		if err != nil {
			log.Fatalf("failed to unmarshal config: %v", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	webhookHost, webhookPortStr, err := net.SplitHostPort(cfg.HostPortAddr)
	if err != nil {
		log.Fatalf("'hostPortAddr' config is malformed: %v", err)
	}

	webhookPort, err := strconv.Atoi(webhookPortStr)
	if err != nil {
		log.Fatalf("'hostPortAddr' port is malformed: %v", err)
	}

	openfgaClient, err := openfgasdk.NewSdkClient(&openfgasdk.ClientConfiguration{
		ApiUrl:  cfg.FGAHostPortAddr,
		StoreId: cfg.FGAStoreID,
	})
	if err != nil {
		log.Fatalf("failed to construct OpenFGA client sdk: %v", err)
	}

	authorizer := &authorizer.OpenFGAAuthorizer{
		OpenFGAClient: openfgaClient,
	}

	k8smanager, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
		//Metrics:
		HealthProbeBindAddress: cfg.HealthProbeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       "92980663.openfga.dev",
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:         webhookHost,
			Port:         webhookPort,
			CertDir:      cfg.HTTPSCertDir,
			CertName:     cfg.HTTPSCertName,
			KeyName:      cfg.HTTPSCertKeyName,
			ClientCAName: cfg.HTTPSCertCAName,
		}),
	})
	if err != nil {
		log.Fatalf("failed to initialize kubernetes manager: %v", err)
	}

	var fgaStore zanzibar.TupleStore

	as := rbacconversion.GetSchema()
	as.Types = append(as.Types, nodeauth.GetSchema().Types...)

	converter := &rbacconversion.GenericConverter{}

	crReconciler := &clusterrolesyncer.ClusterRoleReconciler{
		Client:        k8smanager.GetClient(),
		Scheme:        k8smanager.GetScheme(),
		RBACConverter: converter,
		Zanzibar:      fgaStore,
		TypeRelation:  &as.Types[3],
	}
	if err := crReconciler.SetupWithManager(k8smanager); err != nil {
		log.Fatalf("failed to register ClusterRoleReconciler: %v", err)
	}

	crbReconciler := &clusterrolebindingsyncer.ClusterRoleBindingReconciler{
		Client:        k8smanager.GetClient(),
		Scheme:        k8smanager.GetScheme(),
		RBACConverter: converter,
		Zanzibar:      fgaStore,
		TypeRelation:  &as.Types[0],
	}
	if err := crbReconciler.SetupWithManager(k8smanager); err != nil {
		log.Fatalf("failed to register ClusterRoleBindingReconciler: %v", err)
	}

	roleReconciler := &rolesyncer.RoleReconciler{
		Client:       k8smanager.GetClient(),
		Scheme:       k8smanager.GetScheme(),
		Zanzibar:     fgaStore,
		TypeRelation: &as.Types[2],
	}
	if err := roleReconciler.SetupWithManager(k8smanager); err != nil {
		log.Fatalf("failed to register RoleReconciler: %v", err)
	}

	roleBindingReconciler := &rolebindingsyncer.RoleBindingReconciler{
		Client:       k8smanager.GetClient(),
		Scheme:       k8smanager.GetScheme(),
		Zanzibar:     fgaStore,
		TypeRelation: &as.Types[1],
	}
	if err := roleBindingReconciler.SetupWithManager(k8smanager); err != nil {
		log.Fatalf("failed to register RoleBindingReconciler: %v", err)
	}

	// Register the webhook server's authorization endpoint. The server will be started at k8smanager.Start
	k8smanager.GetWebhookServer().Register(
		"/authorize",
		webhookauthz.NewOpenFGAWebhookAuthorizer(authorizer),
	)

	if err := k8smanager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("/healthz registration failed: %v", err)
	}

	if err := k8smanager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("/readyz registration failed: %v", err)
	}

	if err := k8smanager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("failed to start kubernetes manager: %v", err)
	}
}
