package main

import (
	"testing"
)

var s = `[
      {
        "dgraph.type": [
          "Implementation"
        ],
        "Implementation.path": "cap.implementation.atlassian.jira.install",
        "Implementation.latestRevision": {
          "ImplementationRevision.revision": "0.4.0",
          "ImplementationRevision.spec": {
            "ImplementationSpec.action": {
              "ImplementationAction.args": "{}",
              "ImplementationAction.runnerInterface": "cap.interface.runner.argo"
            },
            "ImplementationSpec.appVersion": "8.x.x"
          },
          "ImplementationRevision.interfaces": [
            {
              "InterfaceRevision.revision": "0.4.1",
              "InterfaceRevision.implementations": [
                {
                  "ImplementationRevision.revision": "0.1.0",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.3.1"
                    },
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.atlassian.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.atlassian.jira",
                    "displayName": "Install Jira"
                  }
                },
                {
                  "ImplementationRevision.revision": "0.4.0",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.atlassian.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.atlassian.jira",
                    "displayName": "Install Jira"
                  }
                },
                {
                  "ImplementationRevision.revision": "0.0.1",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.voltron.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.voltron.jira",
                    "displayName": "Install Jira"
                  }
                }
              ]
            }
          ],
          "ImplementationRevision.metadata": {
            "path": "cap.implementation.atlassian.jira.install",
            "displayName": "Install Jira",
            "description": "Action which installs Jira via Helm chart",
            "maintainers": [
              {
                "Maintainer.email": "team-dev@projectvoltron.dev",
                "Maintainer.url": "https://projectvoltron.dev",
                "Maintainer.name": "Voltron Dev Team"
              }
            ],
            "supportURL": " https://mox.sh/helm",
            "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
            "name": "install",
            "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
            "prefix": "cap.implementation.atlassian.jira"
          }
        },
        "Implementation.revisions": [
          {
            "ImplementationRevision.spec": {
              "ImplementationSpec.action": {
                "ImplementationAction.runnerInterface": "cap.interface.runner.argo",
                "ImplementationAction.args": "{}"
              },
              "ImplementationSpec.appVersion": "8.x.x"
            },
            "ImplementationRevision.interfaces": [
              {
                "InterfaceRevision.revision": "0.3.1",
                "InterfaceRevision.implementations": [
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.1.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.3.1"
                      },
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  }
                ]
              },
              {
                "InterfaceRevision.revision": "0.4.1",
                "InterfaceRevision.implementations": [
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.1.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.3.1"
                      },
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.4.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.voltron.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.voltron.jira"
                    },
                    "ImplementationRevision.revision": "0.0.1",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  }
                ]
              }
            ],
            "ImplementationRevision.metadata": {
              "path": "cap.implementation.atlassian.jira.install",
              "displayName": "Install Jira",
              "description": "Action which installs Jira via Helm chart",
              "maintainers": [
                {
                  "Maintainer.name": "Voltron Dev Team",
                  "Maintainer.email": "team-dev@projectvoltron.dev",
                  "Maintainer.url": "https://projectvoltron.dev"
                }
              ],
              "supportURL": " https://mox.sh/helm",
              "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
              "name": "install",
              "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
              "prefix": "cap.implementation.atlassian.jira"
            },
            "ImplementationRevision.revision": "0.1.0"
          },
          {
            "ImplementationRevision.spec": {
              "ImplementationSpec.action": {
                "ImplementationAction.runnerInterface": "cap.interface.runner.argo",
                "ImplementationAction.args": "{}"
              },
              "ImplementationSpec.appVersion": "8.x.x"
            },
            "ImplementationRevision.interfaces": [
              {
                "InterfaceRevision.revision": "0.4.1",
                "InterfaceRevision.implementations": [
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.1.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.3.1"
                      },
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.4.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.voltron.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.voltron.jira"
                    },
                    "ImplementationRevision.revision": "0.0.1",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  }
                ]
              }
            ],
            "ImplementationRevision.metadata": {
              "path": "cap.implementation.atlassian.jira.install",
              "displayName": "Install Jira",
              "description": "Action which installs Jira via Helm chart",
              "maintainers": [
                {
                  "Maintainer.name": "Voltron Dev Team",
                  "Maintainer.email": "team-dev@projectvoltron.dev",
                  "Maintainer.url": "https://projectvoltron.dev"
                }
              ],
              "supportURL": " https://mox.sh/helm",
              "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
              "name": "install",
              "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
              "prefix": "cap.implementation.atlassian.jira"
            },
            "ImplementationRevision.revision": "0.4.0"
          }
        ],
        "Implementation.name": "install",
        "Implementation.prefix": "cap.implementation.atlassian.jira"
      },
      {
        "dgraph.type": [
          "Implementation"
        ],
        "Implementation.path": "cap.implementation.voltron.jira.install",
        "Implementation.latestRevision": {
          "ImplementationRevision.revision": "0.0.1",
          "ImplementationRevision.spec": {
            "ImplementationSpec.action": {
              "ImplementationAction.args": "{}",
              "ImplementationAction.runnerInterface": "cap.interface.runner.argo"
            },
            "ImplementationSpec.appVersion": "8.x.x"
          },
          "ImplementationRevision.interfaces": [
            {
              "InterfaceRevision.revision": "0.4.1",
              "InterfaceRevision.implementations": [
                {
                  "ImplementationRevision.revision": "0.1.0",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.3.1"
                    },
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.atlassian.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.atlassian.jira",
                    "displayName": "Install Jira"
                  }
                },
                {
                  "ImplementationRevision.revision": "0.4.0",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.atlassian.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.atlassian.jira",
                    "displayName": "Install Jira"
                  }
                },
                {
                  "ImplementationRevision.revision": "0.0.1",
                  "ImplementationRevision.spec": {
                    "ImplementationSpec.appVersion": "8.x.x"
                  },
                  "ImplementationRevision.interfaces": [
                    {
                      "InterfaceRevision.revision": "0.4.1"
                    }
                  ],
                  "ImplementationRevision.metadata": {
                    "path": "cap.implementation.voltron.jira.install",
                    "description": "Action which installs Jira via Helm chart",
                    "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                    "supportURL": " https://mox.sh/helm",
                    "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                    "name": "install",
                    "prefix": "cap.implementation.voltron.jira",
                    "displayName": "Install Jira"
                  }
                }
              ]
            }
          ],
          "ImplementationRevision.metadata": {
            "path": "cap.implementation.voltron.jira.install",
            "displayName": "Install Jira",
            "description": "Action which installs Jira via Helm chart",
            "maintainers": [
              {
                "Maintainer.email": "team-dev@projectvoltron.dev",
                "Maintainer.url": "https://projectvoltron.dev",
                "Maintainer.name": "Voltron Dev Team"
              }
            ],
            "supportURL": " https://mox.sh/helm",
            "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
            "name": "install",
            "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
            "prefix": "cap.implementation.voltron.jira"
          }
        },
        "Implementation.revisions": [
          {
            "ImplementationRevision.spec": {
              "ImplementationSpec.action": {
                "ImplementationAction.runnerInterface": "cap.interface.runner.argo",
                "ImplementationAction.args": "{}"
              },
              "ImplementationSpec.appVersion": "8.x.x"
            },
            "ImplementationRevision.interfaces": [
              {
                "InterfaceRevision.revision": "0.4.1",
                "InterfaceRevision.implementations": [
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.1.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.3.1"
                      },
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.atlassian.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.atlassian.jira"
                    },
                    "ImplementationRevision.revision": "0.4.0",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  },
                  {
                    "ImplementationRevision.metadata": {
                      "description": "Action which installs Jira via Helm chart",
                      "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
                      "supportURL": " https://mox.sh/helm",
                      "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
                      "path": "cap.implementation.voltron.jira.install",
                      "displayName": "Install Jira",
                      "name": "install",
                      "prefix": "cap.implementation.voltron.jira"
                    },
                    "ImplementationRevision.revision": "0.0.1",
                    "ImplementationRevision.spec": {
                      "ImplementationSpec.appVersion": "8.x.x"
                    },
                    "ImplementationRevision.interfaces": [
                      {
                        "InterfaceRevision.revision": "0.4.1"
                      }
                    ]
                  }
                ]
              }
            ],
            "ImplementationRevision.metadata": {
              "path": "cap.implementation.voltron.jira.install",
              "displayName": "Install Jira",
              "description": "Action which installs Jira via Helm chart",
              "maintainers": [
                {
                  "Maintainer.name": "Voltron Dev Team",
                  "Maintainer.email": "team-dev@projectvoltron.dev",
                  "Maintainer.url": "https://projectvoltron.dev"
                }
              ],
              "supportURL": " https://mox.sh/helm",
              "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
              "name": "install",
              "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
              "prefix": "cap.implementation.voltron.jira"
            },
            "ImplementationRevision.revision": "0.0.1"
          }
        ],
        "Implementation.name": "install",
        "Implementation.prefix": "cap.implementation.voltron.jira"
      }
    ]`

var zOut string

func BenchmarkRemoveTypePrefixesFromJSONKeys(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = removeTypePrefixesFromJSONKeys(s)
	}
	zOut = r
}
