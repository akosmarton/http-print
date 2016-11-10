package main

import (
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

const tpl = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
		<h2>HTTP Print Server</h2>
		<p>
			<p>Jobs:</p>
			<table border=1>
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
		</p>
		<p>Last fetch: {{.LastFetchTime}} {{.LastFetchRemoteAddr}}</p>
	</body>
	<script type="text/javascript">
		setTimeout(function(){
	   	window.location.reload(1);
			}, 2000);
	</script>
</html>`

func webRoot(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("tpl").Parse(tpl)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("jobs"))

		type dataJob struct {
			ID          string
			Timestamp   string
			ContentType string
			Len         uint64
		}

		data := struct {
			Jobs                []dataJob
			LastFetchTime       time.Time
			LastFetchRemoteAddr string
		}{
			LastFetchTime:       apiState.LastFetchTime,
			LastFetchRemoteAddr: r.RemoteAddr,
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
