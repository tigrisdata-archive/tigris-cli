package {{.PackageName}}.controller;

import com.tigrisdata.db.client.error.TigrisException;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@RestControllerAdvice
public class TigrisExceptionAdvice {

  @ResponseBody
  @ExceptionHandler(TigrisException.class)
  @ResponseStatus(HttpStatus.INTERNAL_SERVER_ERROR)
  String exceptionHandler(TigrisException ex) {
    String errorMessage = ex.getMessage();
    if (ex.getCause() != null && ex.getCause().getMessage() != null) {
      errorMessage += ", " + ex.getCause().getMessage();
    }
    return errorMessage;
  }
}
