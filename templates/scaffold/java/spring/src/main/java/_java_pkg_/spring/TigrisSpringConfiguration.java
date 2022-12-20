package {{.PackageName}}.spring;

import com.tigrisdata.db.client.StandardTigrisClient;
import com.tigrisdata.db.client.TigrisClient;
import com.tigrisdata.db.client.TigrisDatabase;
import com.tigrisdata.db.client.config.TigrisConfiguration;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class TigrisSpringConfiguration {
  @Bean
  public TigrisDatabase tigrisDatabase(
      @Value("${tigris.db.name}") String dbName, TigrisClient client) {
    return client.getDatabase(dbName);
  }

  @Bean
  public TigrisClient tigrisClient(
      @Value("${tigris.server.url}") String serverURL,
      @Value("${tigris.network.usePlainText:false}") boolean usePlainText) {
    TigrisConfiguration.NetworkConfig.Builder networkConfigBuilder =
            TigrisConfiguration.NetworkConfig.newBuilder();
    if (usePlainText) {
      networkConfigBuilder.usePlainText();
    }
    TigrisConfiguration configuration =
            TigrisConfiguration.newBuilder(serverURL)
            .withNetwork(networkConfigBuilder.build())
            .build();
    return StandardTigrisClient.getInstance(configuration);
  }

  @Bean
  public TigrisInitializer tigrisInitializr(
          TigrisClient tigrisClient, @Value("${tigris.db.name}") String dbName) {
    return new TigrisInitializer(tigrisClient, dbName);
  }
}
