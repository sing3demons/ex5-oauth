package repository

import (
	"context"
	"oauth2-server/database"
	"oauth2-server/models"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func setupSSOSessionTestDB(t *testing.T) (*database.Database, *SSOSessionRepository, func()) {
	// Connect to test database
	db, err := database.Connect("mongodb://localhost:27017", "oauth2_test_sso_sessions")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	repo := NewSSOSessionRepository(db.DB)

	// Cleanup function
	cleanup := func() {
		ctx := context.Background()
		repo.collection.Drop(ctx)
		db.Close()
	}

	// Clear collection before tests
	repo.collection.Drop(context.Background())

	return db, repo, cleanup
}

func TestSSOSessionRepository_Create(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	session := &models.SSOSession{
		SessionID:     "test-session-123",
		UserID:        "user-123",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
	}

	err := repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify CreatedAt and LastActivity were set
	if session.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set automatically")
	}
	if session.LastActivity.IsZero() {
		t.Error("LastActivity should be set automatically")
	}

	// Verify session can be retrieved
	retrieved, err := repo.FindBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	if retrieved.SessionID != session.SessionID {
		t.Errorf("Expected SessionID %s, got %s", session.SessionID, retrieved.SessionID)
	}
	if retrieved.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, retrieved.UserID)
	}
	if !retrieved.Authenticated {
		t.Error("Expected Authenticated to be true")
	}
}

func TestSSOSessionRepository_FindBySessionID(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test finding non-existent session
	_, err := repo.FindBySessionID(ctx, "non-existent")
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected ErrNoDocuments, got %v", err)
	}

	// Create a session
	session := &models.SSOSession{
		SessionID:     "find-test-session",
		UserID:        "user-456",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}
	repo.Create(ctx, session)

	// Test finding existing session
	found, err := repo.FindBySessionID(ctx, "find-test-session")
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if found.SessionID != session.SessionID {
		t.Errorf("Expected SessionID %s, got %s", session.SessionID, found.SessionID)
	}
	if found.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, found.UserID)
	}
}

func TestSSOSessionRepository_UpdateLastActivity(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a session
	session := &models.SSOSession{
		SessionID:     "activity-test-session",
		UserID:        "user-789",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}
	repo.Create(ctx, session)

	// Get initial last activity
	initial, _ := repo.FindBySessionID(ctx, session.SessionID)
	initialActivity := initial.LastActivity

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update last activity
	err := repo.UpdateLastActivity(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to update last activity: %v", err)
	}

	// Verify last activity was updated
	updated, _ := repo.FindBySessionID(ctx, session.SessionID)
	if !updated.LastActivity.After(initialActivity) {
		t.Error("LastActivity should be updated to a later time")
	}
}

func TestSSOSessionRepository_Delete(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a session
	session := &models.SSOSession{
		SessionID:     "delete-test-session",
		UserID:        "user-delete",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}
	repo.Create(ctx, session)

	// Verify session exists
	_, err := repo.FindBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatal("Session should exist before deletion")
	}

	// Delete session
	err = repo.Delete(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session no longer exists
	_, err = repo.FindBySessionID(ctx, session.SessionID)
	if err != mongo.ErrNoDocuments {
		t.Error("Session should not exist after deletion")
	}
}

func TestSSOSessionRepository_DeleteExpired(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create expired sessions
	expiredSession1 := &models.SSOSession{
		SessionID:     "expired-1",
		UserID:        "user-exp-1",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	expiredSession2 := &models.SSOSession{
		SessionID:     "expired-2",
		UserID:        "user-exp-2",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(-2 * time.Hour), // Expired 2 hours ago
	}

	// Create active session
	activeSession := &models.SSOSession{
		SessionID:     "active-1",
		UserID:        "user-active",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // Expires in 7 days
	}

	repo.Create(ctx, expiredSession1)
	repo.Create(ctx, expiredSession2)
	repo.Create(ctx, activeSession)

	// Delete expired sessions
	count, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to delete expired sessions: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 expired sessions deleted, got %d", count)
	}

	// Verify expired sessions are gone
	_, err = repo.FindBySessionID(ctx, "expired-1")
	if err != mongo.ErrNoDocuments {
		t.Error("Expired session 1 should be deleted")
	}

	_, err = repo.FindBySessionID(ctx, "expired-2")
	if err != mongo.ErrNoDocuments {
		t.Error("Expired session 2 should be deleted")
	}

	// Verify active session still exists
	_, err = repo.FindBySessionID(ctx, "active-1")
	if err != nil {
		t.Error("Active session should still exist")
	}
}

func TestSSOSessionRepository_FindByUserID(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	userID := "user-multi-session"

	// Create multiple sessions for the same user
	session1 := &models.SSOSession{
		SessionID:     "user-session-1",
		UserID:        userID,
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}
	session2 := &models.SSOSession{
		SessionID:     "user-session-2",
		UserID:        userID,
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}

	// Create session for different user
	otherSession := &models.SSOSession{
		SessionID:     "other-user-session",
		UserID:        "other-user",
		Authenticated: true,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}

	repo.Create(ctx, session1)
	repo.Create(ctx, session2)
	repo.Create(ctx, otherSession)

	// Find sessions by user ID
	sessions, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find sessions by user ID: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions for user, got %d", len(sessions))
	}

	// Verify all sessions belong to the correct user
	for _, s := range sessions {
		if s.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, s.UserID)
		}
	}

	// Test with non-existent user
	emptySessions, err := repo.FindByUserID(ctx, "non-existent-user")
	if err != nil {
		t.Fatalf("FindByUserID should not error for non-existent user: %v", err)
	}
	if len(emptySessions) != 0 {
		t.Errorf("Expected 0 sessions for non-existent user, got %d", len(emptySessions))
	}
}

func TestSSOSessionRepository_EdgeCases(t *testing.T) {
	_, repo, cleanup := setupSSOSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Create session with zero timestamps", func(t *testing.T) {
		session := &models.SSOSession{
			SessionID:     "zero-timestamp-session",
			UserID:        "user-zero",
			Authenticated: true,
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			// CreatedAt and LastActivity are zero
		}

		err := repo.Create(ctx, session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify timestamps were set
		retrieved, _ := repo.FindBySessionID(ctx, session.SessionID)
		if retrieved.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set automatically")
		}
		if retrieved.LastActivity.IsZero() {
			t.Error("LastActivity should be set automatically")
		}
	})

	t.Run("Update last activity for non-existent session", func(t *testing.T) {
		err := repo.UpdateLastActivity(ctx, "non-existent-session")
		// Should not error, just no-op
		if err != nil {
			t.Errorf("UpdateLastActivity should not error for non-existent session: %v", err)
		}
	})

	t.Run("Delete non-existent session", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-session")
		// Should not error, just no-op
		if err != nil {
			t.Errorf("Delete should not error for non-existent session: %v", err)
		}
	})
}
