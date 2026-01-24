package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployManager_New(t *testing.T) {
	dm := NewDeployManager(nil, nil)
	require.NotNil(t, dm)
	assert.NotNil(t, dm.activeFlows)
	assert.NotNil(t, dm.deploymentLog)
}

func TestDeployManager_ValidateRequest(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	tests := []struct {
		name    string
		req     DeployRequest
		wantErr bool
	}{
		{
			name: "valid full deployment",
			req: DeployRequest{
				Mode: DeployModeFull,
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{ID: "node1", Type: "inject", Name: "Inject"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid modified deployment",
			req: DeployRequest{
				Mode: DeployModeModified,
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{ID: "node1", Type: "inject", Name: "Inject"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid flow deployment",
			req: DeployRequest{
				Mode:   DeployModeFlow,
				FlowID: "flow1",
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{ID: "node1", Type: "inject", Name: "Inject"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			req: DeployRequest{
				Mode: "invalid",
				Flows: []FlowSpec{
					{ID: "flow1", Name: "Test Flow"},
				},
			},
			wantErr: true,
		},
		{
			name: "flow mode without flowId",
			req: DeployRequest{
				Mode: DeployModeFlow,
				Flows: []FlowSpec{
					{ID: "flow1", Name: "Test Flow"},
				},
			},
			wantErr: true,
		},
		{
			name: "no flows",
			req: DeployRequest{
				Mode:  DeployModeFull,
				Flows: []FlowSpec{},
			},
			wantErr: true,
		},
		{
			name: "flow without ID",
			req: DeployRequest{
				Mode: DeployModeFull,
				Flows: []FlowSpec{
					{Name: "Test Flow"},
				},
			},
			wantErr: true,
		},
		{
			name: "flow without name",
			req: DeployRequest{
				Mode: DeployModeFull,
				Flows: []FlowSpec{
					{ID: "flow1"},
				},
			},
			wantErr: true,
		},
		{
			name: "node without ID",
			req: DeployRequest{
				Mode: DeployModeFull,
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{Type: "inject", Name: "Inject"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "node without type",
			req: DeployRequest{
				Mode: DeployModeFull,
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{ID: "node1", Name: "Inject"},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dm.validateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeployManager_DeployModes(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	tests := []struct {
		name string
		mode DeployMode
	}{
		{"full deployment", DeployModeFull},
		{"modified deployment", DeployModeModified},
		{"flow deployment", DeployModeFlow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := DeployRequest{
				Mode: tt.mode,
				Flows: []FlowSpec{
					{
						ID:   "flow1",
						Name: "Test Flow",
						Nodes: []NodeSpec{
							{ID: "node1", Type: "inject", Name: "Inject"},
						},
					},
				},
			}

			if tt.mode == DeployModeFlow {
				req.FlowID = "flow1"
			}

			ctx := context.Background()
			result, err := dm.Deploy(ctx, req)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotZero(t, result.Duration)
			assert.NotZero(t, result.Timestamp)
		})
	}
}

func TestDeployManager_DeployResult(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Test Flow 1",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
			{
				ID:   "flow2",
				Name: "Test Flow 2",
				Nodes: []NodeSpec{
					{ID: "node2", Type: "debug", Name: "Debug"},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := dm.Deploy(ctx, req)
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Message)
	assert.NotZero(t, result.Duration)
	assert.NotZero(t, result.Timestamp)
	assert.NotNil(t, result.DeployedFlows)
	assert.NotNil(t, result.StoppedFlows)
	assert.NotNil(t, result.Errors)
}

func TestDeployManager_DisabledFlows(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:       "flow1",
				Name:     "Enabled Flow",
				Disabled: false,
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
			{
				ID:       "flow2",
				Name:     "Disabled Flow",
				Disabled: true,
				Nodes: []NodeSpec{
					{ID: "node2", Type: "debug", Name: "Debug"},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := dm.Deploy(ctx, req)
	require.NoError(t, err)

	// Disabled flows should not be deployed
	activeFlows := dm.GetActiveFlows()
	assert.NotContains(t, activeFlows, "flow2")
}

func TestDeployManager_DeploymentLog(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Test Flow",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()

	// Deploy multiple times
	for i := 0; i < 3; i++ {
		_, err := dm.Deploy(ctx, req)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Check deployment log
	log := dm.GetDeploymentLog()
	assert.Len(t, log, 3)

	// Check last deployment
	last := dm.GetLastDeployment()
	assert.NotNil(t, last)
	assert.Equal(t, log[len(log)-1].Timestamp, last.Timestamp)
}

func TestDeployManager_GetActiveFlows(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Flow 1",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
			{
				ID:   "flow2",
				Name: "Flow 2",
				Nodes: []NodeSpec{
					{ID: "node2", Type: "debug", Name: "Debug"},
				},
			},
		},
	}

	ctx := context.Background()
	_, err := dm.Deploy(ctx, req)
	require.NoError(t, err)

	activeFlows := dm.GetActiveFlows()
	assert.Len(t, activeFlows, 2)
	assert.Contains(t, activeFlows, "flow1")
	assert.Contains(t, activeFlows, "flow2")
}

func TestDeployManager_StopAll(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Flow 1",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()
	_, err := dm.Deploy(ctx, req)
	require.NoError(t, err)

	// Verify flow is active
	activeFlows := dm.GetActiveFlows()
	assert.Len(t, activeFlows, 1)

	// Stop all
	err = dm.StopAll(ctx)
	assert.NoError(t, err)

	// Verify no active flows
	activeFlows = dm.GetActiveFlows()
	assert.Len(t, activeFlows, 0)
}

func TestDeployManager_RedeployFlow(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	// Initial deployment
	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Flow 1",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()
	result1, err := dm.Deploy(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, result1.DeployedFlows, "flow1")

	// Redeploy same flow
	result2, err := dm.Deploy(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, result2.DeployedFlows, "flow1")
	assert.Contains(t, result2.StoppedFlows, "flow1")
}

func TestDeployManager_FlowModeSpecificFlow(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	// Deploy multiple flows in full mode
	req1 := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Flow 1",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
			{
				ID:   "flow2",
				Name: "Flow 2",
				Nodes: []NodeSpec{
					{ID: "node2", Type: "debug", Name: "Debug"},
				},
			},
		},
	}

	ctx := context.Background()
	_, err := dm.Deploy(ctx, req1)
	require.NoError(t, err)

	// Redeploy only flow1 in flow mode
	req2 := DeployRequest{
		Mode:   DeployModeFlow,
		FlowID: "flow1",
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Flow 1 Updated",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject Updated"},
				},
			},
		},
	}

	result, err := dm.Deploy(ctx, req2)
	require.NoError(t, err)
	assert.Contains(t, result.DeployedFlows, "flow1")
	assert.Len(t, result.DeployedFlows, 1)

	// Verify flow2 is still active
	activeFlows := dm.GetActiveFlows()
	assert.Contains(t, activeFlows, "flow2")
}

func TestDeployManager_DeploymentVersion(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode:    DeployModeFull,
		Version: "v1.2.3",
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Test Flow",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := dm.Deploy(ctx, req)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestDeployManager_ConcurrentDeploys(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Test Flow",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()

	// Run concurrent deployments
	const numDeploys = 5
	done := make(chan bool, numDeploys)

	for i := 0; i < numDeploys; i++ {
		go func() {
			_, err := dm.Deploy(ctx, req)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < numDeploys; i++ {
		<-done
	}

	// Verify deployment log
	log := dm.GetDeploymentLog()
	assert.Len(t, log, numDeploys)
}

func TestDeployManager_LogTruncation(t *testing.T) {
	dm := NewDeployManager(nil, nil)

	req := DeployRequest{
		Mode: DeployModeFull,
		Flows: []FlowSpec{
			{
				ID:   "flow1",
				Name: "Test Flow",
				Nodes: []NodeSpec{
					{ID: "node1", Type: "inject", Name: "Inject"},
				},
			},
		},
	}

	ctx := context.Background()

	// Deploy more than 100 times to test log truncation
	for i := 0; i < 105; i++ {
		_, err := dm.Deploy(ctx, req)
		require.NoError(t, err)
	}

	// Log should be truncated to 100 entries
	log := dm.GetDeploymentLog()
	assert.LessOrEqual(t, len(log), 100)
}
