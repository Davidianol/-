package books.java_main_service.exception;

import lombok.Getter;
import org.springframework.http.HttpStatus;

@Getter
public class HttpStatusException extends RuntimeException {

    private final HttpStatus httpStatus;

    public HttpStatusException(HttpStatus httpStatus, String message) {
        super(message);
        this.httpStatus = httpStatus;
    }
}
