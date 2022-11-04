package {{.PackageName}}.spring;

import com.tigrisdata.db.client.TigrisClient;
import com.tigrisdata.db.client.TigrisDatabase;
import {{.PackageName}}.collections.Order;
import {{.PackageName}}.collections.Product;
import {{.PackageName}}.collections.User;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;

public class TigrisInitializer implements CommandLineRunner {

  private final TigrisClient tigrisClient;
  private final String dbName;

  private static final Logger log = LoggerFactory.getLogger(TigrisInitializer.class);

  public TigrisInitializer(TigrisClient tigrisClient, String dbName) {
    this.tigrisClient = tigrisClient;
    this.dbName = dbName;
  }

  @Override
  public void run(String... args) throws Exception {
    log.info("createDbIfNotExists db: {}", dbName);
    TigrisDatabase tigrisDatabase = tigrisClient.createDatabaseIfNotExists(dbName);
    log.info("creating collections on db {}", dbName);
    tigrisDatabase.createOrUpdateCollections(
{{- $first := true -}}
{{- range .Collections}}
    {{- if $first}}{{$first = false}}{{else}},{{end}}
            {{.Name}}.class
{{- end}}
    );
    log.info("Finished initializing Tigris");
  }
}
