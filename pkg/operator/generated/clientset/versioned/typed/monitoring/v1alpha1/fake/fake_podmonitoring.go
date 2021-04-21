// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/google/gpe-collector/pkg/operator/apis/monitoring/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePodMonitorings implements PodMonitoringInterface
type FakePodMonitorings struct {
	Fake *FakeMonitoringV1alpha1
	ns   string
}

var podmonitoringsResource = schema.GroupVersionResource{Group: "monitoring.googleapis.com", Version: "v1alpha1", Resource: "podmonitorings"}

var podmonitoringsKind = schema.GroupVersionKind{Group: "monitoring.googleapis.com", Version: "v1alpha1", Kind: "PodMonitoring"}

// Get takes name of the podMonitoring, and returns the corresponding podMonitoring object, and an error if there is any.
func (c *FakePodMonitorings) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PodMonitoring, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podmonitoringsResource, c.ns, name), &v1alpha1.PodMonitoring{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodMonitoring), err
}

// List takes label and field selectors, and returns the list of PodMonitorings that match those selectors.
func (c *FakePodMonitorings) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PodMonitoringList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podmonitoringsResource, podmonitoringsKind, c.ns, opts), &v1alpha1.PodMonitoringList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PodMonitoringList{ListMeta: obj.(*v1alpha1.PodMonitoringList).ListMeta}
	for _, item := range obj.(*v1alpha1.PodMonitoringList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podMonitorings.
func (c *FakePodMonitorings) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podmonitoringsResource, c.ns, opts))

}

// Create takes the representation of a podMonitoring and creates it.  Returns the server's representation of the podMonitoring, and an error, if there is any.
func (c *FakePodMonitorings) Create(ctx context.Context, podMonitoring *v1alpha1.PodMonitoring, opts v1.CreateOptions) (result *v1alpha1.PodMonitoring, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podmonitoringsResource, c.ns, podMonitoring), &v1alpha1.PodMonitoring{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodMonitoring), err
}

// Update takes the representation of a podMonitoring and updates it. Returns the server's representation of the podMonitoring, and an error, if there is any.
func (c *FakePodMonitorings) Update(ctx context.Context, podMonitoring *v1alpha1.PodMonitoring, opts v1.UpdateOptions) (result *v1alpha1.PodMonitoring, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podmonitoringsResource, c.ns, podMonitoring), &v1alpha1.PodMonitoring{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodMonitoring), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePodMonitorings) UpdateStatus(ctx context.Context, podMonitoring *v1alpha1.PodMonitoring, opts v1.UpdateOptions) (*v1alpha1.PodMonitoring, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podmonitoringsResource, "status", c.ns, podMonitoring), &v1alpha1.PodMonitoring{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodMonitoring), err
}

// Delete takes name of the podMonitoring and deletes it. Returns an error if one occurs.
func (c *FakePodMonitorings) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(podmonitoringsResource, c.ns, name), &v1alpha1.PodMonitoring{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodMonitorings) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podmonitoringsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PodMonitoringList{})
	return err
}

// Patch applies the patch and returns the patched podMonitoring.
func (c *FakePodMonitorings) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PodMonitoring, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podmonitoringsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PodMonitoring{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodMonitoring), err
}
