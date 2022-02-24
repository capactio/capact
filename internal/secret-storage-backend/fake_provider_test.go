package secretstoragebackend_test

import (
	"sort"

	tellercore "github.com/spectralops/teller/pkg/core"
	tellerutils "github.com/spectralops/teller/pkg/utils"
)

var _ tellercore.Provider = &fakeProvider{}

type fakeProvider struct {
	secrets map[string]map[string]string
}

func (f *fakeProvider) Name() string {
	return "fake"
}

func (f *fakeProvider) GetMapping(kp tellercore.KeyPath) ([]tellercore.EnvEntry, error) {
	kvs := f.getSecret(kp)

	var entries []tellercore.EnvEntry
	for k, v := range kvs {
		entries = append(entries, kp.FoundWithKey(k, v))
	}
	sort.Sort(tellercore.EntriesByKey(entries))
	return entries, nil
}

func (f *fakeProvider) Get(kp tellercore.KeyPath) (*tellercore.EnvEntry, error) {
	kvs := f.getSecret(kp)

	k := kp.EffectiveKey()
	val, ok := kvs[k]
	if !ok {
		ent := kp.Missing()
		return &ent, nil
	}

	ent := kp.Found(val)
	return &ent, nil
}

func (f *fakeProvider) PutMapping(kp tellercore.KeyPath, m map[string]string) error {
	secrets := f.getSecret(kp)

	tellerutils.Merge(m, secrets)

	f.secrets[kp.Path] = secrets

	return nil
}

func (f *fakeProvider) Put(kp tellercore.KeyPath, val string) error {
	k := kp.EffectiveKey()
	return f.PutMapping(kp, map[string]string{k: val})
}

func (f *fakeProvider) Delete(kp tellercore.KeyPath) error {
	kvs := f.getSecret(kp)

	k := kp.EffectiveKey()
	delete(kvs, k)

	if len(kvs) == 0 {
		return f.DeleteMapping(kp)
	}

	f.secrets[kp.Path] = kvs
	return nil
}

func (f *fakeProvider) DeleteMapping(kp tellercore.KeyPath) error {
	delete(f.secrets, kp.Path)
	return nil
}

func (f *fakeProvider) getSecret(kp tellercore.KeyPath) map[string]string {
	secret, ok := f.secrets[kp.Path]
	if !ok {
		return map[string]string{}
	}

	return secret
}
