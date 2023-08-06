package java_test

import (
	"strconv"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/java"
)

func TestDetermineTargetExt_Unsupported(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	ext := java.DetermineTargetExt(fs)

	assert.Equal(t, "jar", ext)
}

func TestDetermineTargetExt_MavenWar(t *testing.T) {
	t.Parallel()

	code := []string{
		`<project>
	<groupId>com.example.projects</groupId>
	<artifactId>documentedproject</artifactId>
	<packaging>war</packaging>
	<version>1.0-SNAPSHOT</version>
	<name>Documented Project</name>
	<url>http://example.com</url>
</project>`,
		// https://stackoverflow.com/q/44297430/12652912
		`<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
		<modelVersion>4.0.0</modelVersion>
		<groupId>MerchantWallet</groupId>
		<artifactId>StellarReceive</artifactId>
		<packaging>war</packaging>
		<version>0.0.1-SNAPSHOT</version>
		<name>StellarReceive Maven Webapp</name>
		<url>http://maven.apache.org</url>
		<properties>
		  <spring.version>4.3.8.RELEASE</spring.version>
		  <jdk.version>1.8</jdk.version>
	  </properties>

	  <dependencies>

			<dependency>
				  <groupId>jstl</groupId>
				  <artifactId>jstl</artifactId>
				  <version>1.2</version>
			  </dependency>

		  <dependency>
		  <groupId>org.springframework</groupId>
		  <artifactId>spring-jdbc</artifactId>
		  <version>${spring.version}</version>
	  </dependency>

		  <dependency>
			  <groupId>org.springframework</groupId>
			  <artifactId>spring-core</artifactId>
			  <version>${spring.version}</version>
		  </dependency>

		  <dependency>
			  <groupId>org.springframework</groupId>
			  <artifactId>spring-web</artifactId>
			  <version>${spring.version}</version>
		  </dependency>

		  <dependency>
			  <groupId>org.springframework</groupId>
			  <artifactId>spring-webmvc</artifactId>
			  <version>${spring.version}</version>
		  </dependency>
	  <!-- hello -->
		  <dependency>
			  <groupId>junit</groupId>
			  <artifactId>junit</artifactId>
			  <version>3.8.1</version>
			  <scope>test</scope>
		  </dependency>

		  <dependency>
		  <groupId>mysql</groupId>
		  <artifactId>mysql-connector-java</artifactId>
		  <version>5.1.6</version>
	  </dependency>


	  </dependencies>

	  <build>
		  <finalName>StellarReceive</finalName>
		  <plugins>

			  <plugin>
				  <groupId>org.apache.maven.plugins</groupId>
				  <artifactId>maven-compiler-plugin</artifactId>
				  <version>3.0</version>
				  <configuration>
					  <source>${jdk.version}</source>
					  <target>${jdk.version}</target>
				  </configuration>
			  </plugin>
		  </plugins>
	  </build>

	  </project>`,
	}

	for i, v := range code {
		v := v
		t.Run("test case #"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			err := afero.WriteFile(fs, "pom.xml", []byte(v), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			ext := java.DetermineTargetExt(fs)
			assert.Equal(t, "war", ext)
		})
	}
}

func TestDetermineTargetExt_MavenNotWar(t *testing.T) {
	t.Parallel()

	code := []string{
		`<project>
	<groupId>com.example.projects</groupId>
	<artifactId>documentedproject</artifactId>
	<version>1.0-SNAPSHOT</version>
	<name>Documented Project</name>
	<url>http://example.com</url>`,
	}

	for i, v := range code {
		v := v
		t.Run("test case #"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			err := afero.WriteFile(fs, "pom.xml", []byte(v), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			ext := java.DetermineTargetExt(fs)
			assert.Equal(t, "jar", ext)
		})
	}
}
