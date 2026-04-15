"""Auth API E2E tests.

Tests cover register, login, refresh, logout, and verify-email endpoints.
Some tests are skipped pending TODO implementations in the backend.
"""

import pytest
from conftest import BASE_URL


class TestRegister:
    """Registration endpoint tests."""

    def test_register_success(self, api_client, random_email):
        """Register a new user with valid credentials."""
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 201
        data = resp.json()
        assert "data" in data

    def test_register_duplicate_email(self, api_client, random_email):
        """Registering the same email twice should return 409."""
        api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 409

    def test_register_weak_password(self, api_client, random_email):
        """Weak passwords should be rejected."""
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "weak",
            "tenant_code": "default",
        })
        assert resp.status_code == 400

    def test_register_missing_fields(self, api_client):
        """Missing required fields should return 400."""
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": "test@example.com",
        })
        assert resp.status_code == 400


class TestLogin:
    """Login endpoint tests."""

    def test_login_success(self, api_client):
        """Login with valid credentials should return tokens."""
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "access_token" in data
        assert "refresh_token" in data
        assert "expires_in" in data

    def test_login_wrong_password(self, api_client):
        """Wrong password should return 401."""
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "WrongPassword1",
            "tenant_code": "default",
        })
        assert resp.status_code == 401

    def test_login_nonexistent_user(self, api_client):
        """Nonexistent user should return 401 (not leak existence)."""
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "nonexistent@example.com",
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 401


class TestTokenRefresh:
    """Token refresh endpoint tests."""

    def test_refresh_token(self, api_client):
        """Valid refresh token should return new token pair."""
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        refresh_token = login_resp.json()["data"]["refresh_token"]

        resp = api_client.post(f"{BASE_URL}/auth/refresh", json={
            "refresh_token": refresh_token,
        })
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "access_token" in data
        assert "refresh_token" in data

    def test_refresh_invalid_token(self, api_client):
        """Invalid refresh token should return 401."""
        resp = api_client.post(f"{BASE_URL}/auth/refresh", json={
            "refresh_token": "invalid-token",
        })
        assert resp.status_code == 401


class TestLogout:
    """Logout endpoint tests."""

    def test_logout(self, api_client, admin_token):
        """Logout should invalidate tokens."""
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        refresh_token = login_resp.json()["data"]["refresh_token"]

        resp = api_client.post(
            f"{BASE_URL}/auth/logout",
            headers=admin_token,
            json={"refresh_token": refresh_token},
        )
        assert resp.status_code == 200


class TestVerifyEmail:
    """Email verification endpoint tests."""

    def test_verify_email_invalid_code(self, api_client, random_email):
        """Invalid verification code should return 400."""
        resp = api_client.post(f"{BASE_URL}/auth/verify-email", json={
            "email": random_email,
            "code": "000000",
        })
        assert resp.status_code == 400

    def test_verify_email_missing_fields(self, api_client):
        """Missing fields should return 400."""
        resp = api_client.post(f"{BASE_URL}/auth/verify-email", json={
            "email": "test@example.com",
        })
        assert resp.status_code == 400
