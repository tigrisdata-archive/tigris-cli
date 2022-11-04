package {{.PackageName}}.collections;

import com.tigrisdata.db.annotation.TigrisField;
import com.tigrisdata.db.annotation.TigrisPrimaryKey;
import com.tigrisdata.db.type.TigrisDocumentCollectionType;
{{with .Collection}}
{{if .HasUUID}}import java.util.UUID;{{end -}}
{{if .HasTime}}import java.util.Date;{{end -}}
import java.util.Objects;
import java.util.Arrays;

{{.Schema}}
{{end}}
