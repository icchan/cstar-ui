// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
"fmt"
"io"
"net/http"
"github.com/gocql/gocql"
"encoding/json"
"strings"
)

func cqlHandler(w http.ResponseWriter, r *http.Request, s *gocql.Session) {
	query := r.FormValue("query")

	fmt.Println("\nq:" + query)

	res, err := s.Query(query).Iter().SliceMap()
	fmt.Printf("\ne: %v", err)
	fmt.Printf("\nr: %v", res)

	enc := json.NewEncoder(w)
	if (err != nil) {
		var m = map[string]string{
			"status": "error",
			"message": err.Error(),
		}

		enc.Encode(m)	
	} else {
		enc.Encode(res)
	}
}

func metaHandler(w http.ResponseWriter, r *http.Request, s *gocql.Session) {

	res, err := s.Query("SELECT keyspace_name, table_name, column_name, kind, type FROM system_schema.columns").Iter().SliceMap()
//	fmt.Printf("\n\ne: %v", err)
//	fmt.Printf("\nr: %v", res)

	enc := json.NewEncoder(w)
	if (err != nil) {
		enc.Encode(err)	
	} else {
		// TODO make this more OO (use objects not maps)

		// init keyspace map
		keyspaces := make(map[string]map[string]map[string]map[string]string)

		for _, row := range res {
			keyspaceName := row["keyspace_name"].(string)
			tableName := row["table_name"].(string)
			columnName := row["column_name"].(string)
			columnType := row["type"].(string)
			columnKind := row["kind"].(string)

			if !strings.HasPrefix(keyspaceName, "system") {		
				// initialize keyspace map if required
				if keyspaces[keyspaceName] == nil {
					keyspaces[keyspaceName] = make(map[string]map[string]map[string]string)
				} 

				// initialise table map if required
				if keyspaces[keyspaceName][tableName] == nil {
					keyspaces[keyspaceName][tableName] = make(map[string]map[string]string)
				} 

				keyspaces[keyspaceName][tableName][columnName] = make(map[string]string)
				keyspaces[keyspaceName][tableName][columnName]["type"] = columnType
				keyspaces[keyspaceName][tableName][columnName]["kind"] = columnKind
			}
		}

		enc.Encode(keyspaces)
	}

}

func main() {

	cassandra := "127.0.0.1"
	listenPort := ":81"

    // connect to the cluster
	cluster := gocql.NewCluster(cassandra)
	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	
	http.HandleFunc("/api/cql", func(w http.ResponseWriter, r *http.Request) {
		cqlHandler(w, r, session)
		})

	http.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		metaHandler(w, r, session)
		})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, frontend)
		})

	fmt.Println("listening on " + listenPort)
	fmt.Println("connected to cassandra at " + cassandra)
	http.ListenAndServe(listenPort, nil)
}


const frontend string = header + middle + footer;

const middle string = `
<div class="form-group">
<label for="comment">CQL Query:</label>
<textarea class="form-control" rows="5" id="comment"></textarea>
</div>
<div>
<button type="button" class="btn btn-default btn-lg">
<span class="glyphicon glyphicon-play" aria-hidden="true"></span> Run Query
</button>
</div>
<div class="alert alert-danger" role="alert">
<span class="glyphicon glyphicon-exclamation-sign" aria-hidden="true"></span>
<span class="sr-only">Error:</span>
<span id="cql-error">Enter a valid email address</span>
</div>

<div>
<table class="table table-striped">
<thead>
<tr>
<th>Row</th>
<th>First Name</th>
<th>Last Name</th>
<th>Email</th>
</tr>
</thead>
<tbody>
<tr>
<td>1</td>
<td>John</td>
<td>Carter</td>
<td>johncarter@mail.com</td>
</tr>
<tr>
<td>2</td>
<td>Peter</td>
<td>Parker</td>
<td>peterparker@mail.com</td>
</tr>
<tr>
<td>3</td>
<td>John</td>
<td>Rambo</td>
<td>johnrambo@mail.com</td>
</tr>
</tbody>
</table>
</div>

<hr />

<div class="checkbox">
<label><input type="checkbox" value="">Option 1</label>
<label><input type="checkbox" value="">Option 2</label>
<label><input type="checkbox" value="">Option 3</label>
</div>

`


const header string = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="description" content="">
<meta name="author" content="">
<link rel="icon" href="favicon.ico">

<title>Starter Template for Bootstrap</title>

<!-- Bootstrap CSS -->
<link href="//maxcdn.bootstrapcdn.com/bootswatch/3.2.0/slate/bootstrap.min.css" rel="stylesheet">
<!--
    // Select one of these: 
<link href="//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css" rel="stylesheet">
<link href="//maxcdn.bootstrapcdn.com/bootswatch/3.2.0/slate/bootstrap.min.css" rel="stylesheet">
<link href="//maxcdn.bootstrapcdn.com/bootswatch/3.2.0/spacelab/bootstrap.min.css" rel="stylesheet">
<link href="//maxcdn.bootstrapcdn.com/bootswatch/3.2.0/amelia/bootstrap.min.css" rel="stylesheet">
-->

<!-- Font Awesome! -->
<link href="//maxcdn.bootstrapcdn.com/font-awesome/4.1.0/css/font-awesome.min.css" rel="stylesheet">

<!-- HTML5 shim and Respond.js IE8 support of HTML5 elements and media queries -->
<!--[if lt IE 9]>
<script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
<script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
<![endif]-->
<style>
body {
	padding-top: 60px;
}
</style>
</head>
<body>
<div class="navbar navbar-inverse navbar-fixed-top" role="navigation">
<div class="container">
<div class="navbar-header">
<a class="navbar-brand" href="#">Project name</a>
</div>
<div class="collapse navbar-collapse">
<ul class="nav navbar-nav">
<li class="active"><a href="#">Home</a></li>
<li><a href="#about">About</a></li>
</ul>
</div><!--/.nav-collapse -->
</div>
</div>

<div class="container">`

const footer string = `
</div><!-- /.container -->

<!-- Bootstrap core JavaScript
================================================== -->
<!-- Placed at the end of the document so the pages load faster -->
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js"></script>
<script src="//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>
</body>
</html>
`