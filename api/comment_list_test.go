package main

import (
	"strings"
	"testing"
	"time"
)

func TestCommentListBasics(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "**foo**", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "**bar**", "approved", time.Now().UTC())

	c, _, err := commentList("temp-commenter-hex", "example.com", "/path.html", false)
	if err != nil {
		t.Errorf("unexpected error listing page comments: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments got %d comments", len(c))
		return
	}

	if c[0].Direction != 0 {
		t.Errorf("expected c.Direction = 0 got c.Direction = %d", c[0].Direction)
		return
	}

	c1Html := strings.TrimSpace(c[1].Html)
	if c1Html != "<p><strong>bar</strong></p>" {
		t.Errorf("expected c[1].Html=[<p><strong>bar</strong></p>] got c[1].Html=[%s]", c1Html)
		return
	}

	c, _, err = commentList(commenterHex, "example.com", "/path.html", false)
	if err != nil {
		t.Errorf("unexpected error listing page comments: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments got %d comments", len(c))
		return
	}

	if c[0].Direction != 0 {
		t.Errorf("expected c.Direction = 1 got c.Direction = %d", c[0].Direction)
		return
	}
}

func TestCommentListEmpty(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	if _, _, err := commentList("temp-commenter-hex", "", "/path.html", false); err == nil {
		t.Errorf("expected error not found listing comments with empty domain")
		return
	}
}

func TestCommentListSelfUnapproved(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())

	c, _, _ := commentList("temp-commenter-hex", "example.com", "/path.html", false)

	if len(c) != 0 {
		t.Errorf("expected user to not see unapproved comment")
		return
	}

	c, _, _ = commentList(commenterHex, "example.com", "/path.html", false)

	if len(c) != 1 {
		t.Errorf("expected user to see unapproved self comment")
		return
	}
}

func TestCommentListAnonymousUnapproved(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commentNew("anonymous", "example.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())

	c, _, _ := commentList("anonymous", "example.com", "/path.html", false)

	if len(c) != 0 {
		t.Errorf("expected user to not see unapproved anonymous comment as anonymous")
		return
	}
}

func TestCommentListIncludeUnapproved(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commentNew("anonymous", "example.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())

	c, _, _ := commentList("anonymous", "example.com", "/path.html", true)

	if len(c) != 1 {
		t.Errorf("expected to see unapproved comments because includeUnapproved was true")
		return
	}
}

func TestCommentListDifferentPaths(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commentNew("anonymous", "example.com", "/path1.html", "root", "**foo**", "unapproved", time.Now().UTC())
	commentNew("anonymous", "example.com", "/path1.html", "root", "**foo**", "unapproved", time.Now().UTC())
	commentNew("anonymous", "example.com", "/path2.html", "root", "**foo**", "unapproved", time.Now().UTC())

	c, _, _ := commentList("anonymous", "example.com", "/path1.html", true)

	if len(c) != 2 {
		t.Errorf("expected len(c) = 2 got len(c) = %d", len(c))
		return
	}

	c, _, _ = commentList("anonymous", "example.com", "/path2.html", true)

	if len(c) != 1 {
		t.Errorf("expected len(c) = 1 got len(c) = %d", len(c))
		return
	}
}

func TestCommentListDifferentDomains(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commentNew("anonymous", "example1.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())
	commentNew("anonymous", "example2.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())

	c, _, _ := commentList("anonymous", "example1.com", "/path.html", true)

	if len(c) != 1 {
		t.Errorf("expected len(c) = 1 got len(c) = %d", len(c))
		return
	}

	c, _, _ = commentList("anonymous", "example2.com", "/path.html", true)

	if len(c) != 1 {
		t.Errorf("expected len(c) = 1 got len(c) = %d", len(c))
		return
	}
}

func TestCommentListAllNoPagination(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "first comment", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "second comment", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "third comment", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 3 {
		t.Errorf("expected 3 comments got %d", len(c))
		return
	}

	if total != 3 {
		t.Errorf("expected totalCount=3 got %d", total)
		return
	}
}

func TestCommentListAllPaginationFirstPage(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	for i := 0; i < 5; i++ {
		commentNew(commenterHex, "example.com", "/path.html", "root", "comment content", "approved", time.Now().UTC())
	}

	pagination := CreatePaginationRequest(1, 2)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments on page 1 got %d", len(c))
		return
	}

	if total != 5 {
		t.Errorf("expected totalCount=5 got %d", total)
		return
	}
}

func TestCommentListAllPaginationSecondPage(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	for i := 0; i < 5; i++ {
		commentNew(commenterHex, "example.com", "/path.html", "root", "comment content", "approved", time.Now().UTC())
	}

	pagination := CreatePaginationRequest(2, 2)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments on page 2 got %d", len(c))
		return
	}

	if total != 5 {
		t.Errorf("expected totalCount=5 got %d", total)
		return
	}
}

func TestCommentListAllPaginationBeyondLastPage(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	for i := 0; i < 3; i++ {
		commentNew(commenterHex, "example.com", "/path.html", "root", "comment content", "approved", time.Now().UTC())
	}

	pagination := CreatePaginationRequest(10, 2)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 0 {
		t.Errorf("expected 0 comments beyond last page got %d", len(c))
		return
	}

	if total != 3 {
		t.Errorf("expected totalCount=3 got %d", total)
		return
	}
}

func TestCommentListAllSearchMatch(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "the quick brown fox", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "lazy dog sleeps", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "another fox story", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "fox"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments matching 'fox' got %d", len(c))
		return
	}

	if total != 2 {
		t.Errorf("expected totalCount=2 got %d", total)
		return
	}
}

func TestCommentListAllSearchNoMatch(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "hello world", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "nonexistent"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 0 {
		t.Errorf("expected 0 comments got %d", len(c))
		return
	}

	if total != 0 {
		t.Errorf("expected totalCount=0 got %d", total)
		return
	}
}

func TestCommentListAllSearchCaseInsensitive(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "Hello World", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "HELLO again", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "goodbye", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "hello"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 case-insensitive matches got %d", len(c))
		return
	}

	if total != 2 {
		t.Errorf("expected totalCount=2 got %d", total)
		return
	}
}

func TestCommentListAllSearchByPath(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/blog/post-one.html", "root", "some content", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/blog/post-two.html", "root", "other content", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/about.html", "root", "about content", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "blog"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments with path matching 'blog' got %d", len(c))
		return
	}

	if total != 2 {
		t.Errorf("expected totalCount=2 got %d", total)
		return
	}
}

func TestCommentListAllSearchWithPagination(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	for i := 0; i < 5; i++ {
		commentNew(commenterHex, "example.com", "/path.html", "root", "matching keyword here", "approved", time.Now().UTC())
	}
	commentNew(commenterHex, "example.com", "/path.html", "root", "no match here", "approved", time.Now().UTC())

	// Page 1 of search results, limit 2
	pagination := CreatePaginationRequest(1, 2)
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "keyword"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments on page 1 of search got %d", len(c))
		return
	}

	if total != 5 {
		t.Errorf("expected totalCount=5 for search got %d", total)
		return
	}

	// Page 3 of search results, limit 2 — should get 1 remaining
	pagination = CreatePaginationRequest(3, 2)
	c, _, total, err = commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "keyword"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 1 {
		t.Errorf("expected 1 comment on page 3 of search got %d", len(c))
		return
	}

	if total != 5 {
		t.Errorf("expected totalCount=5 for search got %d", total)
		return
	}
}

func TestCommentListAllEmptyDomain(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	pagination := CreatePaginationRequest(1, 20)
	if _, _, _, err := commentListAll("", pagination, CommentListAllOptions{IncludeUnapproved: true}); err == nil {
		t.Errorf("expected error for empty domain")
		return
	}
}

func TestCommentListAllExcludeUnapproved(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "approved comment", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "unapproved comment", "unapproved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)

	// includeUnapproved = false
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 1 {
		t.Errorf("expected 1 approved comment got %d", len(c))
		return
	}

	if total != 1 {
		t.Errorf("expected totalCount=1 got %d", total)
		return
	}

	// includeUnapproved = true
	c, _, total, err = commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 comments with unapproved included got %d", len(c))
		return
	}

	if total != 2 {
		t.Errorf("expected totalCount=2 got %d", total)
		return
	}
}

func TestCommentListAllSearchFilterWithUnapproved(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "http://example.com/photo.jpg", "google", "")

	commentNew(commenterHex, "example.com", "/path.html", "root", "approved fox", "approved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "unapproved fox", "unapproved", time.Now().UTC())
	commentNew(commenterHex, "example.com", "/path.html", "root", "approved dog", "approved", time.Now().UTC())

	pagination := CreatePaginationRequest(1, 100)

	// Search for "fox" excluding unapproved
	c, _, total, err := commentListAll("example.com", pagination, CommentListAllOptions{Search: "fox"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 1 {
		t.Errorf("expected 1 approved fox comment got %d", len(c))
		return
	}

	if total != 1 {
		t.Errorf("expected totalCount=1 got %d", total)
		return
	}

	// Search for "fox" including unapproved
	c, _, total, err = commentListAll("example.com", pagination, CommentListAllOptions{IncludeUnapproved: true, Search: "fox"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(c) != 2 {
		t.Errorf("expected 2 fox comments with unapproved got %d", len(c))
		return
	}

	if total != 2 {
		t.Errorf("expected totalCount=2 got %d", total)
		return
	}
}
