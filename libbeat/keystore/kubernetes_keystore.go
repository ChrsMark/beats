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

package keystore

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/bus"
	"github.com/elastic/beats/v7/libbeat/logp"
)

type KubernetesKeystores map[string]Keystore

type KubernetesKeystoresRegistry struct {
	kubernetesKeystores KubernetesKeystores
	logger              *logp.Logger
	client              k8s.Interface
}

// KubernetesSecretsKeystore allows to retrieve passwords from Kubernetes secrets for a given namespace.
type KubernetesSecretsKeystore struct {
	namespace string
	client    k8s.Interface
}

// Factoryk8s Create the right keystore with the configured options.
func Factoryk8s(keystoreNamespace string, ks8client k8s.Interface) (Keystore, error) {
	keystore, err := NewKubernetesSecretsKeystore(keystoreNamespace, ks8client)
	return keystore, err
}

func NewKubernetesKeystoresRegistry(logger *logp.Logger, client k8s.Interface) *KubernetesKeystoresRegistry {
	return &KubernetesKeystoresRegistry{
		kubernetesKeystores: KubernetesKeystores{},
		logger:              logger,
		client:              client,
	}
}

// GetK8sKeystore return a KubernetesSecretsKeystore if it already exists for a given namespace of creates a new one.
func (kr *KubernetesKeystoresRegistry) GetK8sKeystore(event bus.Event) Keystore {
	namespace := ""
	if val, ok := event["kubernetes"]; ok {
		kubernetesMeta := val.(common.MapStr)
		ns, err := kubernetesMeta.GetValue("namespace")
		if err != nil {
			kr.logger.Debugf("Cannot retrieve kubernetes namespace from event: %s", event)
			return nil
		}
		namespace = ns.(string)
	}
	if namespace != "" {
		// either retrieve already stored keystore or create a new one for the namespace
		storedKeystore := kr.lookupForKeystore(namespace)
		if storedKeystore != nil {
			return storedKeystore
		}
		k8sKeystore, _ := Factoryk8s(namespace, kr.client)
		kr.kubernetesKeystores["namespace"] = k8sKeystore
		return k8sKeystore
	}
	return nil
}

func (kr *KubernetesKeystoresRegistry) lookupForKeystore(keystoreNamespace string) Keystore {
	if keystore, ok := kr.kubernetesKeystores[keystoreNamespace]; ok {
		return keystore
	}
	return nil
}

// NewKubernetesSecretsKeystore returns an new k8s Keystore
func NewKubernetesSecretsKeystore(keystoreNamespace string, ks8client k8s.Interface) (Keystore, error) {
	keystore := KubernetesSecretsKeystore{
		namespace: keystoreNamespace,
		client:    ks8client,
	}
	return &keystore, nil
}

// Retrieve return a SecureString instance that will contains both the key and the secret.
func (k *KubernetesSecretsKeystore) Retrieve(key string) (*SecureString, error) {
	// key = "kubernetes:somenamespace:somesecret:value"
	toks := strings.Split(key, ".")
	ns := toks[1]
	secretName := toks[2]
	secretVar := toks[3]
	if ns != k.namespace {
		return nil, fmt.Errorf("cannot access Kubernetes secrets from a different namespace than: %v", ns)
	}
	secretIntefrace := k.client.CoreV1().Secrets(ns)
	secrets, err := secretIntefrace.List(metav1.ListOptions{})
	if err != nil {
		// log the error here
		return nil, err
	}
	if len(secrets.Items) == 0 {
		// log the error here
		return nil, fmt.Errorf("no secrets found for namespace: %v", ns)
	}
	secret, err := secretIntefrace.Get(secretName, metav1.GetOptions{})
	if err != nil {
		// log the error here
		return nil, err
	}
	secretString := secret.Data[secretVar]
	return NewSecureString(secretString), nil
}

// GetConfig returns common.Config representation of the key / secret pair to be merged with other
// loaded configuration.
func (k *KubernetesSecretsKeystore) GetConfig() (*common.Config, error) {
	return nil, nil
}

// IsPersisted return if the keystore is physically persisted on disk.
func (k *KubernetesSecretsKeystore) IsPersisted() bool {
	return true
}
