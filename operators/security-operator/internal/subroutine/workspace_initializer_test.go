package subroutine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.platform-mesh.io/security-operator/internal/subroutine/mocks"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kcpcorev1alpha1 "github.com/kcp-dev/sdk/apis/core/v1alpha1"
)

func TestWorkspaceInitializer_Terminate(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		setupMocks  func(*mocks.MockClient)
		expectGet   bool
		expectError bool
	}{
		{
			name:      "success - deletes the org store",
			path:      "root:orgs:test-workspace",
			expectGet: true,
			setupMocks: func(m *mocks.MockClient) {
				m.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*v1alpha1.Store"), mock.Anything).Return(nil).Once()
			},
		},
		{
			name:      "success - already gone is not an error",
			path:      "root:orgs:test-workspace",
			expectGet: true,
			setupMocks: func(m *mocks.MockClient) {
				m.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*v1alpha1.Store"), mock.Anything).
					Return(apierrors.NewNotFound(schema.GroupResource{Group: "core.platform-mesh.io", Resource: "stores"}, "test-workspace")).Once()
			},
		},
		{
			name:      "error - delete fails",
			path:      "root:orgs:test-workspace",
			expectGet: true,
			setupMocks: func(m *mocks.MockClient) {
				m.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*v1alpha1.Store"), mock.Anything).
					Return(assert.AnError).Once()
			},
			expectError: true,
		},
		{
			name:       "no-op - empty path yields no store name",
			path:       "",
			expectGet:  false,
			setupMocks: func(m *mocks.MockClient) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			kcpHelper := mocks.NewMockKCPClientGetter(t)
			tt.setupMocks(mockClient)
			if tt.expectGet {
				kcpHelper.EXPECT().NewClientForLogicalCluster(mock.Anything, "root:orgs").Return(mockClient, nil).Once()
			}

			w := &workspaceInitializer{kcpClientGetter: kcpHelper}
			lc := &kcpcorev1alpha1.LogicalCluster{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kcp.io/path": tt.path}},
			}

			_, err := w.Terminate(context.Background(), lc)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
