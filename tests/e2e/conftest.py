import os
import uuid

import pytest
import requests

BASE_URL = os.getenv("IAM_BASE_URL", "http://localhost:8080/api/v1")

ADMIN_EMAIL = "admin@iam.local"
ADMIN_PASSWORD = "Admin@123"
DEFAULT_TENANT_CODE = "default"


@pytest.fixture
def api_client():
    """HTTP client with session management."""
    session = requests.Session()
    session.headers.update({"Content-Type": "application/json"})
    return session


@pytest.fixture
def admin_token(api_client):
    """Login as admin and return auth headers."""
    resp = api_client.post(f"{BASE_URL}/auth/login", json={
        "email": ADMIN_EMAIL,
        "password": ADMIN_PASSWORD,
        "tenant_code": DEFAULT_TENANT_CODE,
    })
    assert resp.status_code == 200, f"admin login failed: {resp.text}"
    data = resp.json()["data"]
    token = data["access_token"]
    return {"Authorization": f"Bearer {token}"}


@pytest.fixture
def admin_refresh_token(api_client):
    """Login as admin and return the refresh token."""
    resp = api_client.post(f"{BASE_URL}/auth/login", json={
        "email": ADMIN_EMAIL,
        "password": ADMIN_PASSWORD,
        "tenant_code": DEFAULT_TENANT_CODE,
    })
    assert resp.status_code == 200
    return resp.json()["data"]["refresh_token"]


@pytest.fixture
def random_email():
    """Generate a random email for registration tests."""
    return f"test_{uuid.uuid4().hex[:8]}@example.com"
