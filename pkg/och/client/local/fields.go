package local

const typeInstanceFields = `
	resourceVersion
	metadata {
	  id
	  attributes {
	    path
	    revision
	  }
	}
	spec {
	  typeRef {
	    path
	    revision
	  }
	  value
	}
	uses {
		metadata {
			id
		}
		spec {
			typeRef {
				path
				revision
			}
		}
	}
	usedBy {
		metadata {
			id
		}
		spec {
			typeRef {
				path
				revision
			}
		}
	}
`
