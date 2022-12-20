package {{.PackageName}}.controller;

import com.tigrisdata.db.client.Filters;
import com.tigrisdata.db.client.TigrisCollection;
import com.tigrisdata.db.client.TigrisDatabase;
import com.tigrisdata.db.client.error.TigrisException;
import {{.PackageName}}.collections.{{.Collection.Name}};
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
{{with .Collection}}
@RestController
@RequestMapping("{{.JSON}}")
public class {{.Name}}Controller {

  private final TigrisCollection<{{.Name}}> {{.NameDecap}}TigrisCollection;

  public {{.Name}}Controller(TigrisDatabase tigrisDatabase) {
    this.{{.NameDecap}}TigrisCollection = tigrisDatabase.getCollection({{.Name}}.class);
  }

  @PostMapping("/")
  public ResponseEntity<String> create(@RequestBody {{.Name}} {{.NameDecap}}) throws TigrisException {
    {{.NameDecap}}TigrisCollection.insert({{.NameDecap}});
    return ResponseEntity.status(HttpStatus.CREATED).body("{{.NameDecap}} created");
  }

  @GetMapping("/{id}")
  public {{.Name}} read(@PathVariable("id") int id) throws TigrisException {
    return {{.NameDecap}}TigrisCollection.readOne(Filters.eq("id", id)).get();
  }

  @DeleteMapping("/{id}")
  public ResponseEntity<String> delete(@PathVariable("id") int id) throws TigrisException {
    {{.NameDecap}}TigrisCollection.delete(Filters.eq("id", id));
    return ResponseEntity.status(HttpStatus.OK).body("{{.NameDecap}} deleted");
  }
}
{{end}}
