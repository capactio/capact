package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/mindstand/gogm"
	"log"
	"net/http"
	"projectvoltron.dev/voltron/poc/graph-db/neo4j/graphql"
)

func main() {
	config := gogm.Config{
		IndexStrategy: gogm.VALIDATE_INDEX, //other options are ASSERT_INDEX and IGNORE_INDEX
		PoolSize:      50,
		Port:          7687,
		IsCluster:     false, //tells it whether or not to use `bolt+routing`
		Host:          "0.0.0.0",
		Password:      "root",
		Username:      "neo4j",
	}

	types := []interface{}{
		&graphql.GenericMetadata{},
		&graphql.Interface{},
		&graphql.InterfaceGroup{},
		&graphql.Signature{},
		&graphql.InterfaceRevision{},
		&graphql.Maintainer{},
	}

	err := gogm.Init(&config, types...)
	if err != nil {
		panic(err)
	}

	//param is readonly, we're going to make stuff so we're going to do read write
	sess, err := gogm.NewSession(false)
	if err != nil {
		panic(err)
	}

	//close the session
	defer sess.Close()

	mustLoadModels(sess)

	gqlCfg := graphql.Config{
		Resolvers: graphql.NewResolver(sess),
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)
	srv := handler.NewDefaultServer(executableSchema)

	http.Handle("/", playground.Handler("Neo4j GoGM PoC", "/graphql"))
	http.Handle("/graphql", srv)

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func mustLoadModels(sess *gogm.Session) {
	err := sess.Begin()
	if err != nil {
		panic(err)
	}

	interfaceGroups := []*graphql.InterfaceGroup{
		{
			Metadata:   &graphql.GenericMetadata{
				Name:              "foo",
				Prefix:            "prefix",
				Path:              "path",
				DisplayName:       "Foo",
				Description:       "Foo",
				Maintainers:       []*graphql.Maintainer{
					{
						Name:            "Maitainer 1",
						Email:           "foo@bar.com",
					},
				},
			},
			Signature:  &graphql.Signature{
				Och:            "och1",
			},
			Interfaces: []*graphql.Interface{
				{
					Name:           "interface 1",
					Prefix:         "prefix int",
					Path:           "int path",
					Revisions:      []*graphql.InterfaceRevision{
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "Revision 1",
							},
							Revision:  "0.0.1",
						},
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "Revision 2",
							},
							Revision:  "0.0.2",
						},
					},
				},
			},
		},

		{
			Metadata:   &graphql.GenericMetadata{
				Name:              "bar",
				Prefix:            "prefix",
				Path:              "path",
				DisplayName:       "Bar",
				Description:       "Bar",
				Maintainers:       []*graphql.Maintainer{
					{
						Name:            "Maitainer 1",
						Email:           "foo@bar.com",
					},
					{
						Name:            "Maitainer 1",
						Email:           "bar@bar.com",
					},
				},
			},
			Signature:  &graphql.Signature{
				Och:            "och2",
			},
			Interfaces: []*graphql.Interface{
				{
					Name:           "bar interface 1",
					Prefix:         "prefix int",
					Path:           "int path",
					Revisions:      []*graphql.InterfaceRevision{
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "bar 1 Revision 1",
							},
							Revision:  "0.0.1",
						},
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "bar 1 Revision 2",
							},
							Revision:  "0.0.2",
						},
					},
				},
				{
					Name:           "bar interface 2",
					Prefix:         "prefix int",
					Path:           "int path",
					Revisions:      []*graphql.InterfaceRevision{
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "bar 2 Revision 1",
							},
							Revision:  "0.0.1",
						},
						{
							Metadata:  &graphql.GenericMetadata{
								Name: "bar 2 Revision 2",
							},
							Revision:  "0.0.2",
						},
					},
				},
			},
		},
	}


	for _, iG := range interfaceGroups {
		err := sess.SaveDepth(iG, 10)
		if err != nil {
			panic(err)
		}
	}

	err = sess.Commit()
	if err != nil {
		panic(err)
	}
}
