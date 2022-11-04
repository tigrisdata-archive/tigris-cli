package {{.PackageName}};

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class {{.DBNameCamel}}Application {

  public static void main(String[] args) {
    new SpringApplication({{.DBNameCamel}}Application.class).run(args);
  }
}
