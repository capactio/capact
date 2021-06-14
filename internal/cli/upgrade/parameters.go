package upgrade

type (
	InputParameters struct {
		Version                string `json:"version"`
		IncreaseResourceLimits bool   `json:"-"`
		Override               struct {
			HelmRepoURL  string       `json:"helmRepoURL"`
			CapactValues CapactValues `json:"capactValues,omitempty"`
			Neo4jValues  Neo4jValues  `json:"neo4jValues,omitempty"`
		} `json:"override"`
	}
	Neo4jValues struct {
		Neo4j struct {
			Core struct {
				Resources Resources `json:"resources,omitempty"`
			} `json:"core,omitempty"`
		} `json:"neo4j,omitempty"`
	}
	CapactValues struct {
		Engine    Engine    `json:"engine,omitempty"`
		Gateway   Gateway   `json:"gateway,omitempty"`
		HubPublic HubPublic `json:"hub-public,omitempty"`
		Global    struct {
			ContainerRegistry struct {
				Tag  string `json:"overrideTag,omitempty"`
				Path string `json:"path,omitempty"`
			} `json:"containerRegistry,omitempty"`
		} `json:"global,omitempty"`
	}
	ResourcesQuantity struct {
		CPU    string `json:"cpu,omitempty"`
		Memory string `json:"memory,omitempty"`
	}
	Resources struct {
		Limits   ResourcesQuantity `json:"limits,omitempty"`
		Requests ResourcesQuantity `json:"requests,omitempty"`
	}
	Gateway struct {
		Resources Resources `json:"resources,omitempty"`
	}
	HubPublic struct {
		Resources Resources `json:"resources,omitempty"`
	}
	Engine struct {
		TestSetup struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"testSetup,omitempty"`
	}
)

func increasedGatewayResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "300m",
			Memory: "128Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "100m",
			Memory: "48Mi",
		},
	}
}

func increasedHubPublicResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "400m",
			Memory: "512Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "200m",
			Memory: "128Mi",
		},
	}
}

func increasedNeo4jResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "1",
			Memory: "3072Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "500m",
			Memory: "1768Mi",
		},
	}
}
