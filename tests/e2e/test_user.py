"""User profile API E2E tests."""

from conftest import BASE_URL


class TestUserProfile:
    """User profile endpoint tests."""

    def test_get_profile(self, api_client, admin_token):
        """Get current user profile."""
        resp = api_client.get(f"{BASE_URL}/users/me", headers=admin_token)
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "email" in data
        assert data["email"] == "admin@iam.local"

    def test_update_profile(self, api_client, admin_token):
        """Update user profile name and avatar."""
        resp = api_client.put(
            f"{BASE_URL}/users/me",
            headers=admin_token,
            json={"name": "Test Admin", "avatar_url": "https://example.com/avatar.png"},
        )
        assert resp.status_code == 200

        # Verify update persisted
        resp = api_client.get(f"{BASE_URL}/users/me", headers=admin_token)
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert data["name"] == "Test Admin"

    def test_change_password(self, api_client, admin_token):
        """Change password and verify new one works."""
        resp = api_client.put(
            f"{BASE_URL}/users/me/password",
            headers=admin_token,
            json={"old_password": "Admin@123", "new_password": "NewAdmin@123"},
        )
        assert resp.status_code == 200

        # Verify new password works
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "NewAdmin@123",
            "tenant_code": "default",
        })
        assert login_resp.status_code == 200
        new_token = login_resp.json()["data"]["access_token"]
        new_headers = {"Authorization": f"Bearer {new_token}"}

        # Reset back to original password
        resp = api_client.put(
            f"{BASE_URL}/users/me/password",
            headers=new_headers,
            json={"old_password": "NewAdmin@123", "new_password": "Admin@123"},
        )
        assert resp.status_code == 200

    def test_change_password_wrong_old(self, api_client, admin_token):
        """Wrong old password should be rejected."""
        resp = api_client.put(
            f"{BASE_URL}/users/me/password",
            headers=admin_token,
            json={"old_password": "WrongPassword1", "new_password": "NewAdmin@123"},
        )
        assert resp.status_code == 400

    def test_get_profile_unauthenticated(self, api_client):
        """Accessing profile without token should return 401."""
        resp = api_client.get(f"{BASE_URL}/users/me")
        assert resp.status_code == 401
