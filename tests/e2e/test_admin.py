"""Admin API E2E tests."""

import uuid

from conftest import BASE_URL


class TestAdminTenants:
    """Admin tenant management tests."""

    def test_list_tenants(self, api_client, admin_token):
        """List all tenants."""
        resp = api_client.get(f"{BASE_URL}/admin/tenants", headers=admin_token)
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "tenants" in data
        assert len(data["tenants"]) >= 1

    def test_create_tenant(self, api_client, admin_token):
        """Create a new tenant."""
        resp = api_client.post(
            f"{BASE_URL}/admin/tenants",
            headers=admin_token,
            json={"name": "Test Tenant", "unique_code": f"test_{uuid.uuid4().hex[:8]}"},
        )
        assert resp.status_code == 201

    def test_create_duplicate_tenant(self, api_client, admin_token):
        """Duplicate tenant code should return 409."""
        resp = api_client.post(
            f"{BASE_URL}/admin/tenants",
            headers=admin_token,
            json={"name": "Duplicate", "unique_code": "default"},
        )
        assert resp.status_code == 409

    def test_get_tenant(self, api_client, admin_token):
        """Get a specific tenant by ID."""
        # Get the default tenant ID from list first
        resp = api_client.get(f"{BASE_URL}/admin/tenants", headers=admin_token)
        assert resp.status_code == 200
        tenant_id = resp.json()["data"]["tenants"][0]["id"]

        resp = api_client.get(f"{BASE_URL}/admin/tenants/{tenant_id}", headers=admin_token)
        assert resp.status_code == 200


class TestAdminRoles:
    """Admin role management tests."""

    def test_list_roles(self, api_client, admin_token):
        """List roles for a tenant."""
        resp = api_client.get(f"{BASE_URL}/admin/roles", headers=admin_token)
        assert resp.status_code == 200

    def test_create_role(self, api_client, admin_token):
        """Create a new role."""
        resp = api_client.post(
            f"{BASE_URL}/admin/roles",
            headers=admin_token,
            json={"name": f"test_role_{uuid.uuid4().hex[:8]}", "description": "Test role"},
        )
        assert resp.status_code == 201


class TestAdminAuditLogs:
    """Admin audit log tests."""

    def test_list_audit_logs(self, api_client, admin_token):
        """List audit logs."""
        resp = api_client.get(f"{BASE_URL}/admin/audit-logs", headers=admin_token)
        assert resp.status_code == 200

    def test_list_audit_logs_with_time_range(self, api_client, admin_token):
        """List audit logs with time range filter."""
        resp = api_client.get(
            f"{BASE_URL}/admin/audit-logs",
            headers=admin_token,
            params={"start_time": "2026-01-01T00:00:00Z", "end_time": "2027-01-01T00:00:00Z"},
        )
        assert resp.status_code == 200


class TestAdminUnauthorized:
    """Admin access without proper role."""

    def test_admin_without_token(self, api_client):
        """Accessing admin without auth token should return 401."""
        resp = api_client.get(f"{BASE_URL}/admin/tenants")
        assert resp.status_code == 401
