package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	// Import godotenv for loading .env.test file
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	fmt.Println("Setting up test environment...")
	
	err := godotenv.Load("../.env.test")
	if err != nil {
		fmt.Printf("Warning: Error loading .env.test file: %v\n", err)
		fmt.Println("Some tests requiring environment variables may fail")
	} else {
		fmt.Println("Successfully loaded .env.test environment variables")
	}
	
	// Get the directory of this file
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	
	// Path to fixtures directory
	fixturesDir := filepath.Join(dir, "fixtures")
	
	// Ensure fixture directory exists
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		err := os.MkdirAll(fixturesDir, 0755)
		if err != nil {
			fmt.Printf("Failed to create fixtures directory: %v\n", err)
			os.Exit(1)
		}
	}
	
	// Create fixture files
	createFixture(fixturesDir, "subreddit.json", subredditJSON)
	createFixture(fixturesDir, "user_about.json", userAboutJSON)
	createFixture(fixturesDir, "post.json", postJSON)
	createFixture(fixturesDir, "user_posts.json", userPostsJSON)
	createFixture(fixturesDir, "user_comments.json", userCommentsJSON)
	createFixture(fixturesDir, "more_comments.json", moreCommentsJSON)
	
	// Print key environment variables for debugging
	printEnvVar("REDDIT_PROXY_URLS")
	printEnvVar("REDDIT_USER_AGENT")
	printEnvVar("TEST_MODE")
	
	// Run all tests
	exitCode := m.Run()
	
	fmt.Println("Cleaning up test environment...")
	
	os.Exit(exitCode)
}

// Helper function to print environment variables
func printEnvVar(name string) {
	value := os.Getenv(name)
	if value != "" {
		// Mask sensitive values
		if name == "REDDIT_PROXY_URLS" {
			fmt.Printf("%s is set (value masked for security)\n", name)
		} else {
			fmt.Printf("%s=%s\n", name, value)
		}
	} else {
		fmt.Printf("%s is not set\n", name)
	}
}

func createFixture(dir, filename, content string) {
	path := filepath.Join(dir, filename)
	// Replace deprecated ioutil.WriteFile with os.WriteFile
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Failed to create fixture %s: %v\n", filename, err)
		os.Exit(1)
	}
}

const subredditJSON = `{
  "data": {
    "children": [
      {
        "kind": "t3",
        "data": {
          "id": "abc123",
          "title": "Test post",
          "selftext": "This is a test post",
          "author": "testuser",
          "score": 42,
          "created_utc": 1620000000,
          "subreddit": "test",
          "permalink": "/r/test/comments/abc123/test_post",
          "url": "https://reddit.com/r/test/comments/abc123/test_post"
        }
      }
    ],
    "after": "t3_next123"
  }
}`

const userAboutJSON = `{
  "data": {
    "name": "testuser",
    "created_utc": 1620000000,
    "link_karma": 100,
    "comment_karma": 200
  }
}`

const postJSON = `[
  {
    "data": {
      "children": [
        {
          "data": {
            "id": "abc123",
            "title": "Test post",
            "author": "testuser",
            "created_utc": 1620000000,
            "score": 42,
            "permalink": "/r/test/comments/abc123/test_post",
            "selftext": "This is a test post"
          }
        }
      ]
    }
  },
  {
    "data": {
      "children": [
        {
          "kind": "t1",
          "data": {
            "id": "comment1",
            "author": "commenter",
            "body": "This is a comment",
            "score": 5,
            "created_utc": 1620000100,
            "replies": ""
          }
        }
      ]
    }
  }
]`

const userPostsJSON = `{
  "data": {
    "children": [
      {
        "kind": "t3",
        "data": {
          "id": "abc123",
          "title": "User post",
          "selftext": "This is a user post",
          "author": "testuser",
          "score": 42,
          "created_utc": 1620000000,
          "subreddit": "test",
          "permalink": "/r/test/comments/abc123/user_post",
          "url": "https://reddit.com/r/test/comments/abc123/user_post"
        }
      }
    ],
    "after": "t3_next123"
  }
}`

const userCommentsJSON = `{
  "data": {
    "children": [
      {
        "kind": "t1",
        "data": {
          "id": "comment1",
          "body": "This is a user comment",
          "author": "testuser",
          "score": 5,
          "created_utc": 1620000100,
          "subreddit": "test",
          "link_id": "t3_abc123",
          "link_title": "Test post"
        }
      }
    ],
    "after": "t1_next123"
  }
}`

const moreCommentsJSON = `{
  "json": {
    "data": {
      "things": [
        {
          "kind": "t1",
          "data": {
            "id": "more1",
            "author": "commenter2",
            "body": "This is a nested comment",
            "score": 3,
            "created_utc": 1620000200,
            "replies": ""
          }
        }
      ]
    }
  }
}`