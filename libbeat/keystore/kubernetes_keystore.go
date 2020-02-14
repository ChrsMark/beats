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
	"github.com/elastic/beats/libbeat/common"
	k8s "k8s.io/client-go/kubernetes"
)

type KubernetesKeystores map[string]KubernetesSecretsKeystore

var kubernetesKeystoresRegistry = KubernetesKeystores{}

// KubernetesSecretsKeystore Allows to retrieve passwords from Kubernetes secrets for a given namespace.
type KubernetesSecretsKeystore struct {
	namespace string
	client k8s.Interface
}

// NewKubernetesSecretsKeystore returns an new k8s Keystore
func NewKubernetesSecretsKeystore(keystoreNamespace string, ks8client k8s.Interface) (Keystore, error) {
	// search for in the keystore registry to find if there is keystore already for that namespace
	storedKeystore, err := lookupForKeystore(keystoreNamespace)
	if err != nil {
		return nil, err
	}
	if storedKeystore != nil {
		return storedKeystore, nil
	}

	keystore := KubernetesSecretsKeystore{
		namespace: keystoreNamespace,
		client:    ks8client,
	}
	err = keystore.load()
	if err != nil {
		return nil, err
	}
	kubernetesKeystoresRegistry[keystoreNamespace] = keystore
	return &keystore, nil
}

func lookupForKeystore(keystoreNamespace string) (Keystore, error){
	if keystore, ok := kubernetesKeystoresRegistry[keystoreNamespace]; ok {
		err := keystore.load()
		if err != nil {
			return nil, err
		} else {
			return &keystore, nil
		}
	}
	return nil, nil
}

// Retrieve return a SecureString instance that will contains both the key and the secret.
func (k *KubernetesSecretsKeystore) Retrieve(key string) (*SecureString, error) {

	// here retrieve from API by using the key which is the name of the Namespace
	return NewSecureString([]byte{}), nil
}

// Store add the key pair to the secret store and mark the store as dirty.
func (k *KubernetesSecretsKeystore) Store(key string, value []byte) error {
	return nil
}

// Delete an existing key from the store and mark the store as dirty.
func (k *KubernetesSecretsKeystore) Delete(key string) error {
	return nil
}

// Save persists the in memory data to disk if needed.
func (k *KubernetesSecretsKeystore) Save() error {
	return nil
}

// List return the availables keys.
func (k *KubernetesSecretsKeystore) List() ([]string, error) {
	return []string{}, nil
}

// GetConfig returns common.Config representation of the key / secret pair to be merged with other
// loaded configuration.
func (k *KubernetesSecretsKeystore) GetConfig() (*common.Config, error) {
}

// Create create an empty keystore, if the store already exist we will return an error.
func (k *KubernetesSecretsKeystore) Create(override bool) error {
}

// IsPersisted return if the keystore is physically persisted on disk.
func (k *KubernetesSecretsKeystore) IsPersisted() bool {
	return true
}

func (k *KubernetesSecretsKeystore) load() error {
	return nil
}


// checkPermission enforces permission on the keystore file itself, the file should have strict
// permission (0600) and the keystore should refuses to start if its not the case.
func (k *KubernetesSecretsKeystore) checkPermissions(f string) error {
	return nil
}
