package main

import (
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

const tpl = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet">
		<title>HTTP Print Server</title>
	</head>
	<body>
		<div class="container">
			<nav class="navbar navbar-default">
				<div class="navbar-header">
					<a class="navbar-brand" href="/">HTTP Print Server</a>
				</div>
			</nav>
		<div class="panel panel-default">
			<div class="panel-heading">Printer {{.Printer}} - Jobs</div>
				<div class="panel-body">
					<table class="table">
						<tr>
							<th>Timestamp</th>
							<th>ID</th>
							<th>Content Type</th>
							<th>Length</th>
						</tr>
						{{range .Jobs }}
						<tr>
							<td>{{ .Timestamp }}</td>
							<td>{{ .ID }}</td>
							<td>{{ .ContentType }}</td>
							<td>{{ .Len }}</td>
						</tr>
						{{end}}
					</table>
				</div>
			</div>
		</div>
	</body>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>
	<script type="text/javascript">
		setTimeout(function(){
	   	window.location.reload(1);
			}, 2000);
	</script>
</html>`

func webRoot(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("tpl").Parse(tpl)

	printer := mux.Vars(r)["printer"]

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(printer))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		type dataJob struct {
			ID          string
			Timestamp   string
			ContentType string
			Len         uint64
		}

		data := struct {
			Printer string
			Jobs    []dataJob
		}{
			Printer: printer,
		}

		b.ForEach(func(k, v []byte) error {
			var j job
			j.GobDecode(v)

			dj := dataJob{
				ID:          hex.EncodeToString(k),
				Timestamp:   time.Unix(int64(j.Timestamp), 0).UTC().String(),
				ContentType: j.ContentType,
				Len:         j.Len,
			}

			data.Jobs = append(data.Jobs, dj)

			return nil
		})
		t.Execute(w, data)
		return nil
	})
}
