FROM gradle:8.12.1-jdk17 AS build
WORKDIR /app/java-main-service

# Копируем gradlew и папку gradle из каталога java-main-service
COPY java-main-service/gradlew .
COPY java-main-service/gradle ./gradle

# Копируем файлы сборки проекта
COPY java-main-service/build.gradle.kts .
COPY java-main-service/settings.gradle.kts .

# Копируем исходники проекта
COPY java-main-service/ ./

# Копируем папку с proto-файлами (на уровень выше, так как build.gradle.kts ищет их по $rootDir/../api)
COPY api/ ../api/

# Делаем gradlew исполняемым
RUN chmod +x gradlew

# Выполняем сборку проекта
RUN ./gradlew build --no-daemon

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/java-main-service/build/libs/*.jar app.jar
ENTRYPOINT ["java", "-jar", "app.jar"]
