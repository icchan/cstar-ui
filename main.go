package main

import (
	"fmt"
	"io"
	"net/http"
	"github.com/gocql/gocql"
	"encoding/json"
	"strings"
	"flag"
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

func metaHandlerV3(w http.ResponseWriter, r *http.Request, s *gocql.Session) {
  res, err := s.Query("select keyspace_name, columnfamily_name, column_name, type, validator from system.schema_columns").Iter().SliceMap()

  enc := json.NewEncoder(w)
  if (err != nil) {
    enc.Encode(err) 
  } else {
    // TODO make this more OO (use objects not maps)

    // init keyspace map
    keyspaces := make(map[string]map[string]map[string]map[string]string)

    for _, row := range res {
      keyspaceName := row["keyspace_name"].(string)
      tableName := row["columnfamily_name"].(string)
      columnName := row["column_name"].(string)
      columnTypeRaw := row["validator"].(string)
      columnKind := row["type"].(string)

      //columnType := columnTypeRaw[32:len(columnTypeRaw)-4]

      columnType1 := strings.Replace(columnTypeRaw, "org.apache.cassandra.db.marshal.", "", -1)
      columnType := strings.Replace(columnType1, "Type", "", -1)

      if (!strings.HasPrefix(keyspaceName, "system") && !strings.HasPrefix(keyspaceName, "dse_") ){   
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

func metaHandlerV4(w http.ResponseWriter, r *http.Request, s *gocql.Session) {
	res, err := s.Query("SELECT keyspace_name, table_name, column_name, kind, type FROM system_schema.columns").Iter().SliceMap()

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

	hostPtr := flag.String("host", "127.0.0.1", "cassandra host")
	listenPtr := flag.String("listen", ":8080", "listen address:port")
	flag.Parse()

	cassandra := *hostPtr
	listenPort := *listenPtr

    // connect to the cluster
	cluster := gocql.NewCluster(cassandra)
	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 3
	session, _ := cluster.CreateSession()
	defer session.Close()

	// static content
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// execute cql
	http.HandleFunc("/api/cql", func(w http.ResponseWriter, r *http.Request) {
		cqlHandler(w, r, session)
		})

	// get info about keyspaces, tables, columns
	http.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		metaHandlerV3(w, r, session)
		})

	// serve the html page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, frontend)
		})

	fmt.Println("listening on " + listenPort)
	fmt.Println("connected to cassandra at " + cassandra)
	err := http.ListenAndServe(listenPort, nil)

  fmt.Printf("should not occur error: %v\n", err);
}


const frontend string = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="">
  <meta name="author" content="">
  <link rel="icon" href="favicon.ico">

  <title>Cassandra Quick Access</title>

  <!-- Bootstrap CSS -->
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/flatly/bootstrap.min.css" rel="stylesheet">

    <!--
    // Select one of these: 
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/flatly/bootstrap.min.css" rel="stylesheet">
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/slate/bootstrap.min.css" rel="stylesheet">
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/superhero/bootstrap.min.css" rel="stylesheet">
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/darkly/bootstrap.min.css" rel="stylesheet">
  <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.3.6/lumen/bootstrap.min.css" rel="stylesheet">

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
        li.meta-columns span {
          display: inline-block;
        }
        li.meta-columns span.cql-col {
          width: 150px;
        }
        li.meta-columns span.cql-type {
          width: 150px;
        }
        li.meta-columns span.cql-kind {
          width: 10px;
        }
        .alert {
            margin-top: 10px;
        }
      </style>
    </head>
    <body>
      <div class="navbar navbar-inverse navbar-fixed-top" role="navigation">
        <div class="container">
          <div class="navbar-header">
            <a class="navbar-brand" href="#">CStar UI</a>
          </div>
          <div class="collapse navbar-collapse">
            <ul class="nav navbar-nav" id="topmenu">
              <li><a href="#tab-cql">CQL</a></li>
              <li><a href="#tab-tables">Tables</a></li>
              <li><a href="#tab-meta">Metadata</a></li>
              <li><a href="#tab-other">Other</a></li>
            </ul>
          </div><!--/.nav-collapse -->
        </div>
      </div>

      <div class="container" id="main">
        <!-- ================================== CQL TAB here ==================================  -->
        <div id="tab-cql">
          <h3>Let's Execute Some CQL</h3>
          <div class="form-group">
            <label for="comment">CQL Query:</label>
            <textarea class="form-control" rows="4" id="textarea-cql">SELECT * FROM testme.users;</textarea>
          </div>

          <div class="btn-toolbar">
            <button type="button" class="btn btn-success" id="btn-cql">
              <span class="glyphicon glyphicon-play" aria-hidden="true"></span> Run Query
            </button>
          </div>

          <div class="alert alert-danger" role="alert" data-bind="visible: VIEW.cqlError">
            <span class="glyphicon glyphicon-exclamation-sign" aria-hidden="true"></span>
            <span class="sr-only">Error:</span>
            <span id="cql-error"><span data-bind="text: VIEW.cqlErrorMsg"></span></span>
          </div>

          <div class="results" id="cql-results" data-bind="visible: VIEW.cqlItems().length > 0">
            <h4>Results</h4>
            <table class="table table-striped">
              <thead>
                <tr data-bind="foreach: {data: VIEW.cqlColumnNames, as: 'cName'}">
                  <th> <span data-bind="text: cName"></span> </th>
                </tr>
              </thead>
              <tbody data-bind="foreach: VIEW.cqlItems">
                <tr data-bind="foreach: $parent.cqlColumnNames">
                    <td data-bind="text: $parent[$data]"></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <!-- ================================== TABLES TAB here ==================================  -->
        <div id="tab-tables">
          <h3>Let's Work With Tables</h3>
          <div class="checkbox">
            <div id="meta-display">
              <div data-bind="foreach: {data: VIEW.META, as: 'keyspace'}">
                <h4>Keyspace: <span data-bind="text: keyspace.name"></span></h4>
                <div data-bind="foreach: { data: keyspace.tables, as: 'table' }">
                  <label><input type="checkbox" data-bind="value: keyspace.name + '.' + table.name" ><span data-bind="text: table.name"></span></label>
                </div>
              </div>
            </div>
          </div>
          <div class="btn-toolbar">
            <button type="button" class="btn btn-success">
              <span class="glyphicon glyphicon-list" aria-hidden="true"></span> Select *
            </button>
            <button type="button" class="btn btn-danger">
              <span class="glyphicon glyphicon-remove" aria-hidden="true"></span> Truncate
            </button>
          </div>

          <div class="results" id="cql-results">
            <h4>Results</h4>
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
              </tbody>
            </table>
          </div>
        </div>
        <!-- ================================== META TAB here ==================================  -->
        <div id="tab-meta">
          <h3>Metadata</h3>
          <p>What the db looks like</p>

          <div id="meta-display">
            <div class="panel panel-warning keyspace-meta" data-bind="foreach: {data: VIEW.META, as: 'keyspace'}">
              <div class="panel-heading"><span data-bind="text: keyspace.name"> </span> <span class="tbl-collapse">[-]</span></div>
              <div class="table-meta panel-body">
                <ul class="list-group" data-bind="foreach: { data: keyspace.tables, as: 'table' }">
                  <li class="list-group-item">
                    <h4><span data-bind="text: table.name"></span> <small>[-]</small></h4>
                    <ul class="list-group" data-bind="foreach: { data: table.columns, as: 'column' }">
                      <li class="list-group-item meta-columns">
                        <b><span class="cql-col"  data-bind="text: column.name"></span></b>
                        <span class="cql-type" data-bind="text: column.type"></span> 
                        <span class="cql-kind"  data-bind="text: column.kind"></span>
                      </li>
                    </ul>
                  </li>
                </ul>
              </div>
            </div>
          </div>

        </div>
        <!-- ================================== OTHER TAB here ==================================  -->
        <div id="tab-other">
          <h3>Other..</h3>
          <p>Some other stuff I haven't thought of yet..</p>
        </div>

      </div><!-- /.container -->


    <!-- Bootstrap core JavaScript
    ================================================== -->
    <!-- Placed at the end of the document so the pages load faster -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.0.0/jquery.min.js"></script>
    <script src="//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/knockout/3.4.0/knockout-min.js"></script>

    <script>
      $(function() {

        $('.panel-heading').click(function() {
          $(this).next('.panel-body').toggle();
          console.log($(this));
        });

//        $(".alert").hide();
//        $(".results").hide();

        // set up top menu and tabs
        initTabMenu();

        // Actual app code
        initCqlTab();
        loadMeta();
      });

      // globals for knockout js
      var VIEW = {
          META: ko.observableArray(),
          cqlItems: ko.observableArray(),
          cqlError: ko.observable(false),
          cqlErrorMsg: ko.observable('some error')
      }

      VIEW.cqlColumnNames = ko.computed(function () {
              var self = VIEW;
              if (self.cqlItems().length === 0)
                  return [];
              var props = [];
              var obj = self.cqlItems()[0];
              for (var name in obj)
                  props.push(name);
              return props;
          });

      ko.applyBindings(VIEW);


      // functions
      function initTabMenu() {
        // set up the tabs
        $("#main").children().hide();
        var type = window.location.hash.substr(1);

        if (type == "") {
          type = "tab-cql";
        }
        $("#" + type).show();

        // set up the menu
        $("#topmenu > li > a").click(function(){
          var thisref = $(this).attr('href');
          console.log(thisref);
          $("#main").children().hide();
          $(thisref).show();
        });
      }

      function initCqlTab(){
        $('#btn-cql').click(function() {
          VIEW.cqlError(false);

          var query = $('#textarea-cql').val();
          // TODO validate it

          console.log(query);

          $.getJSON( "/api/cql?query=" + query, function( data ) {
            console.log(data);

            // TODO validate
            if (data.status != null) {
              console.log("CQL failed");
              VIEW.cqlError(true);
              VIEW.cqlErrorMsg(data.message);
              console.log(data.message);
            } else {

              VIEW.cqlItems.removeAll();

              var arrayLength = data.length;
              for (var i = 0; i < arrayLength; i++) {
                var row = data[i];
                VIEW.cqlItems.push(row);
              }

            }

          });
        });

      }


      function loadMeta() {
          $.getJSON( "/api/meta", function( data ) {
            console.log(data);
            displayMeta(data);
          });
      }

      function displayMeta(metaResult) {
          //console.log(metaResult);

          // TODO ensure the input is valid before continuing
          VIEW.META.removeAll();

          // extract keyspaces
          for (var ksName in metaResult) {
            if (metaResult.hasOwnProperty(ksName)) {

              var tblRes = metaResult[ksName];
              var tblList = [];

              // extract tables
              for (var tblName in tblRes) {
                var colRes = tblRes[tblName];
                var colList = [];

                // extract column data
                for (var colName in colRes) {
                  var colMeta = colRes[colName];

                  var dKind = "";
                  if (colMeta.kind =="partition_key" ) {
                    dKind = "K";
                  } else if ((colMeta.kind =="clustering" ) || (colMeta.kind =="clustering_key" ) ){
                    dKind = "C";
                  } else if (colMeta.kind =="static" ) {
                    dKind = "S";
                  }

                  var column = {
                    "name": colName,
                    "type": colMeta.type,
                    "kind": dKind
                  }
                  colList.push(column)
                }

                colList.sort(function(a, b){
                  if (a.kind == "") {
                    return 1;
                  } else {
                    // TODO return a.kind.localeCompare(b.kind);
                    return -1;               
                  }
                });

                var table = {
                  "name": tblName,
                  "columns": colList
                }

                tblList.push(table);
              }

              var keyspace = {
                "name" : ksName,
                "tables" : tblList
              };

              VIEW.META.push(keyspace);
            }
          }

        }

    </script>

  </body>
  </html>
`