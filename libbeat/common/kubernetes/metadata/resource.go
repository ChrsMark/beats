// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package metadata

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/common/safemapstr"
)

// Resource generates metadata for any kubernetes resource
type Resource struct {
	config     *Config
	clusterURL string
}

// NewResourceMetadataGenerator creates a metadata generator for a generic resource
func NewResourceMetadataGenerator(cfg *common.Config, client k8s.Interface) *Resource {
	var config Config
	config.Unmarshal(cfg)

	return &Resource{
		config:     &config,
		clusterURL: getClusterURL(client),
	}
}

// Generate generates pod metadata from a resource object
func (r *Resource) Generate(kind string, obj kubernetes.Resource, opts ...FieldOptions) common.MapStr {
	return common.MapStr{
		"kubernetes": r.GenerateK8s(kind, obj, opts...),
	}
}

// Generate generates pod metadata from a resource object
func (r *Resource) GenerateECS(obj kubernetes.Resource) common.MapStr {
	ecsMeta := common.MapStr{}
	if r.clusterURL != "" {
		ecsMeta.Put("orchestrator.cluster.url", r.clusterURL)
	}
	return ecsMeta
}

// Generate takes a kind and an object and creates metadata for the same
func (r *Resource) GenerateK8s(kind string, obj kubernetes.Resource, options ...FieldOptions) common.MapStr {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil
	}

	labelMap := common.MapStr{}
	if len(r.config.IncludeLabels) == 0 {
		labelMap = GenerateMap(accessor.GetLabels(), r.config.LabelsDedot)
	} else {
		labelMap = generateMapSubset(accessor.GetLabels(), r.config.IncludeLabels, r.config.LabelsDedot)
	}

	// Exclude any labels that are present in the exclude_labels config
	for _, label := range r.config.ExcludeLabels {
		labelMap.Delete(label)
	}

	annotationsMap := generateMapSubset(accessor.GetAnnotations(), r.config.IncludeAnnotations, r.config.AnnotationsDedot)

	meta := common.MapStr{
		strings.ToLower(kind): common.MapStr{
			"name": accessor.GetName(),
			"uid":  string(accessor.GetUID()),
		},
	}

	if accessor.GetNamespace() != "" {
		// TODO make this namespace.name in 8.0
		safemapstr.Put(meta, "namespace", accessor.GetNamespace())
	}

	// Add controller metadata if present
	if r.config.IncludeCreatorMetadata {
		for _, ref := range accessor.GetOwnerReferences() {
			if ref.Controller != nil && *ref.Controller {
				switch ref.Kind {
				// TODO grow this list as we keep adding more `state_*` metricsets
				case "Deployment",
					"ReplicaSet",
					"StatefulSet":
					safemapstr.Put(meta, strings.ToLower(ref.Kind)+".name", ref.Name)
				}
			}
		}
	}

	if len(labelMap) != 0 {
		safemapstr.Put(meta, "labels", labelMap)
	}

	if len(annotationsMap) != 0 {
		safemapstr.Put(meta, "annotations", annotationsMap)
	}

	for _, option := range options {
		option(meta)
	}

	return meta
}

func generateMapSubset(input map[string]string, keys []string, dedot bool) common.MapStr {
	output := common.MapStr{}
	if input == nil {
		return output
	}

	for _, key := range keys {
		value, ok := input[key]
		if ok {
			if dedot {
				dedotKey := common.DeDot(key)
				output.Put(dedotKey, value)
			} else {
				safemapstr.Put(output, key, value)
			}
		}
	}

	return output
}

func GenerateMap(input map[string]string, dedot bool) common.MapStr {
	output := common.MapStr{}
	if input == nil {
		return output
	}

	for k, v := range input {
		if dedot {
			label := common.DeDot(k)
			output.Put(label, v)
		} else {
			safemapstr.Put(output, k, v)
		}
	}

	return output
}

func getClusterURL(client k8s.Interface) string {
	endpoint, err := client.CoreV1().Endpoints("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return ""
	}
	if len(endpoint.Subsets) == 0 {
		return ""
	}
	subset := endpoint.Subsets[0]
	if len(subset.Addresses) == 0 || len(subset.Ports) == 0 {
		return ""
	}
	ip := subset.Addresses[0].IP
	port := subset.Ports[0].Port
	if ip == "" || port == int32(0) {
		return ""
	}
	return fmt.Sprintf("%v:%v", ip, port)
}
