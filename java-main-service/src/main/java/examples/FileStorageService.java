package examples;

import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;

@Service
public class FileStorageService {
        private final Path booksStorageLocation = Paths.get("/app/books");
        private final Path usersStorageLocation = Paths.get("/app/users");

        public void saveBookFile(MultipartFile file, String bookId) throws IOException {
                Files.createDirectories(booksStorageLocation);
                Path targetLocation = booksStorageLocation.resolve(bookId + ".pdf");
                Files.copy(file.getInputStream(), targetLocation, StandardCopyOption.REPLACE_EXISTING);
        }

        public void saveUserAvatar(MultipartFile file, String userId) throws IOException {
                Files.createDirectories(usersStorageLocation);
                Path targetLocation = usersStorageLocation.resolve(userId + ".jpg");
                Files.copy(file.getInputStream(), targetLocation, StandardCopyOption.REPLACE_EXISTING);
        }
}
