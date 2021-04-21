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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	monitoringv1alpha1 "github.com/google/gpe-collector/pkg/operator/apis/monitoring/v1alpha1"
	versioned "github.com/google/gpe-collector/pkg/operator/generated/clientset/versioned"
	internalinterfaces "github.com/google/gpe-collector/pkg/operator/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/google/gpe-collector/pkg/operator/generated/listers/monitoring/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// PodMonitoringInformer provides access to a shared informer and lister for
// PodMonitorings.
type PodMonitoringInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.PodMonitoringLister
}

type podMonitoringInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPodMonitoringInformer constructs a new informer for PodMonitoring type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPodMonitoringInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPodMonitoringInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPodMonitoringInformer constructs a new informer for PodMonitoring type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPodMonitoringInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MonitoringV1alpha1().PodMonitorings(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MonitoringV1alpha1().PodMonitorings(namespace).Watch(context.TODO(), options)
			},
		},
		&monitoringv1alpha1.PodMonitoring{},
		resyncPeriod,
		indexers,
	)
}

func (f *podMonitoringInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPodMonitoringInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *podMonitoringInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&monitoringv1alpha1.PodMonitoring{}, f.defaultInformer)
}

func (f *podMonitoringInformer) Lister() v1alpha1.PodMonitoringLister {
	return v1alpha1.NewPodMonitoringLister(f.Informer().GetIndexer())
}
