package capact

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"capact.io/capact/pkg/httputil"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

type HelmHelper struct {
	HTTPClient *http.Client
}

func NewHelmHelper() *HelmHelper {
	httpClient := httputil.NewClient(30 * time.Second)
	return &HelmHelper{
		HTTPClient: httpClient,
	}
}

// GetLatestVersion loads an index file and returns version of the latest chart:
//	- for the @latest Helm charts sort by Created field
//  - for all others repos sort by SemVer
//
// Assumption that all charts are versioned in the same way.
func (h *HelmHelper) GetLatestVersion(repoURL string, chart string) (string, error) {
	// by default sort by SemVer, so even if someone pushed bugfix of older
	// release we will not take it.
	sortFn := func(in *repo.IndexFile) {
		in.SortEntries()
	}

	// `main` (@latest) charts are versioned via SHA commit, so we need to sort
	// them via Created time.
	if repoURL == HelmRepoLatest {
		sortFn = func(in *repo.IndexFile) {
			sort.Sort(ByCreatedTime(in.Entries[chart]))
		}
	}

	url := fmt.Sprintf("%s/index.yaml", repoURL)
	resp, err := h.HTTPClient.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "while getting capactio Helm Chart repository index.yaml")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	i := &repo.IndexFile{}
	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return "", errors.Wrapf(err, "Index fetch from %q is malformed", url)
	}

	sortFn(i)

	capactEntry, ok := i.Entries[chart]
	if !ok {
		return "", fmt.Errorf("no entry %q in Helm Chart repository index.yaml", chart)
	}

	if len(capactEntry) == 0 {
		return "", fmt.Errorf("no Chart versions for entry %q in Helm Chart repository index.yaml", chart)
	}

	return capactEntry[0].Version, nil
}

type ByCreatedTime repo.ChartVersions

func (b ByCreatedTime) Len() int           { return len(b) }
func (b ByCreatedTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByCreatedTime) Less(i, j int) bool { return b[i].Created.After(b[j].Created) }

func ValuesFromString(values string) (map[string]interface{}, error) {
	v := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(values), &v)
	return v, err
}
