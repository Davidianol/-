import com.google.protobuf.gradle.*


plugins {
    java
    id("org.springframework.boot") version "3.4.5"
    id("io.spring.dependency-management") version "1.1.7"
    id("com.google.protobuf") version "0.9.4"
    id("antlr")
}

group = "books"
version = "0.0.1-SNAPSHOT"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(17)
    }
}

configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}

repositories {
    mavenCentral()
}

sourceSets["main"].java.srcDir("build/generated/source/proto/main/java")
dependencies {
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-jdbc")
    implementation("org.hibernate.orm:hibernate-core:6.3.2.Final")
    implementation("org.springdoc:springdoc-openapi-starter-webmvc-ui:2.8.5")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    implementation("org.springframework.boot:spring-boot-starter-web")
    compileOnly("org.projectlombok:lombok")
    runtimeOnly("org.postgresql:postgresql:42.7.5")
    annotationProcessor("org.projectlombok:lombok")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
    implementation("org.hibernate.validator:hibernate-validator:7.0.5.Final")

    developmentOnly("org.springframework.boot:spring-boot-devtools")
    developmentOnly("org.springframework.boot:spring-boot-docker-compose")

    implementation("com.google.protobuf:protobuf-java:3.25.5")
    implementation("io.grpc:grpc-netty-shaded:1.51.0")
    implementation("io.grpc:grpc-protobuf:1.51.0")
    implementation("javax.annotation:javax.annotation-api:1.3.2")
    implementation("io.grpc:grpc-stub:1.51.0")
    implementation("com.google.guava:guava:32.1.2-jre")
    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")

    testImplementation("org.testcontainers:postgresql:1.17.6")
    testImplementation("org.testcontainers:junit-jupiter:1.17.6")

    implementation("io.jsonwebtoken:jjwt-api:0.11.5")
    runtimeOnly("io.jsonwebtoken:jjwt-impl:0.11.5")
    runtimeOnly("io.jsonwebtoken:jjwt-jackson:0.11.5")

    antlr("org.antlr:antlr4:4.13.0")
    implementation("org.antlr:antlr4-runtime:4.13.0")


}
protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:3.25.5"
    }
    plugins {
        id("grpc") {
            artifact = "io.grpc:protoc-gen-grpc-java:1.51.0"
        }
    }
    sourceSets {
        main {
            proto {
                srcDir("$rootDir/../api")
            }
        }
    }
    generateProtoTasks {
        all().forEach { task ->
            task.plugins {
                id("grpc")
            }
        }
    }
}
tasks.withType<Test> {
    useJUnitPlatform()
}
