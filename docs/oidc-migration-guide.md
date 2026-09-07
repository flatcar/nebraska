# OIDC Migration Guide

Migration guide for Nebraska's secure OIDC implementation with Authorization Code Flow + PKCE.

## What's Changed
- Frontend handles OIDC flow directly (no backend client secret)
- PKCE security for SPA authentication
- Stateless backend with JWT validation
- Access tokens must carry the configured API audience (see below)

## Upgrading an existing OIDC deployment

Access token audience validation is a **breaking change** for OIDC deployments. Deployments using `--auth-mode=noop` or `--auth-mode=github` are unaffected.

| Your deployment | What happens on upgrade |
|---|---|
| `--oidc-audience` not set | Nebraska **refuses to start** until you set it |
| `--oidc-audience` set, and your provider puts that value in the access token `aud` | works, no action needed |
| `--oidc-audience` set, but your provider does **not** put it in `aud` | Nebraska starts normally, then **every request fails with 401** |
| the API audience is also mapped into **ID tokens**, and clients send the ID token | those requests now fail with 401; send the access token instead |

The third row is the one to watch. **Setting the flag is not enough. Your provider has to actually issue access tokens carrying that audience.** Keycloak does not do this by default, and it does not support the `audience` request parameter, so only an audience protocol mapper works.

### Recommended order (no downtime)

This change on the provider side works with older Nebraska versions too, so apply it first and verify before upgrading:

1. Configure your provider to put the API audience in access tokens (step 1 below). Existing Nebraska versions ignore the extra audience, so nothing breaks.
2. Check a freshly issued access token really contains it:
   ```bash
   echo "<access-token>" | cut -d. -f2 | tr '_-' '/+' | base64 -d | jq .aud
   ```
3. Upgrade Nebraska and set `--oidc-audience` to the same value.

If you need to upgrade before you can change the provider, start with `--oidc-skip-audience-check`. Nebraska will start and log a warning, skipping **audience** validation until you finish the migration (tokens that clearly identify as ID tokens are still rejected).

## Migration Steps

### 1. OIDC Provider Configuration

**Required Changes:**
1. Change client type: **Confidential** → **Public** (SPA)
2. Remove client secret
3. Set redirect URI: `https://your-domain.com/auth/callback`
4. Enable CORS for your Nebraska domain

**Provider Examples:**
- **Keycloak:** 
  - Change "Access Type" to "public"
  - Authentication flow should be set to standard
  - Add to "Valid Redirect URIs": `https://your-domain.com/auth/callback`
  - Set "Web Origins": `https://your-domain.com`
  - Set "Post logout redirect URI": `https://your-domain.com/`
  - Add an Audience protocol mapper to the Nebraska client or a dedicated client scope:
    - Included Custom Audience: a backend API identifier such as `nebraska-api`
    - Add to access token: enabled
    - Add to ID token: disabled
- **Auth0:** 
  - Change Application Type to "Single Page Application"  
  - Add to "Allowed Callback URLs": `https://your-domain.com/auth/callback`
  - Add to "Allowed Web Origins": `https://your-domain.com`
  - Add "Allowed Logout URLs": `http://localhost:8000/`
  - **Important:** Disable "Implicit" grant type, keep only "Authorization Code"
  - Create an API in Auth0 Dashboard → APIs → Create API
    - Set an identifier (e.g., `https://nebraska-api` - doesn't need to be a real URL)
    - Use this identifier as the audience parameter in Nebraska config
- **Okta:** 
  - Change to "SPA" application type. Grant type should be Authorization Code.
  - Use a Custom Authorization Server and configure its Audience as the Nebraska API identifier.
  - Add to "Sign-in redirect URIs": `https://your-domain.com/auth/callback`
  - Add to "Trusted Origins": `https://your-domain.com`
  - Sign-out redirect URIs: `http://localhost:8000`
  - Set CORS under Trusted Origins
- **Azure AD:**
  - Set Platform to "Single-page application"
  - Add to "Redirect URIs": `https://your-domain.com/auth/callback`
  - Ensure the redirect URI is set correctly
  - Under Implicit grant and hybrid flows, ensure both checkboxes are unchecked
  - Configure Logout URL: http://localhost:8000
  - For CORS, go to Expose an API
  - Expose/register the Nebraska backend API and use its application ID/URI as the audience
- **Dex:**
  - Dex has no separate resource server, so the access token audience is the client ID itself. Set `--oidc-audience` to the same value as `--oidc-client-id`.
  - Dex marks neither token kind and gives its access tokens and ID tokens the same audience and the same claims, so Nebraska cannot tell them apart. A Dex ID token is accepted wherever an access token is, so treat the two as equally sensitive.

**How Nebraska validates it:** the `aud` claim of the access token must contain the value of `--oidc-audience`, as required by [RFC 9068](https://www.rfc-editor.org/rfc/rfc9068) section 4. A token the provider issued for a different application is rejected even though its signature and issuer are valid.

Keep the API audience out of ID tokens. In the Keycloak audience mapper this is the default ("Add to ID token" is off). Nebraska also rejects a token that clearly says it is an ID token. A token whose header type is `at+jwt` is accepted as an access token. A token that carries a Keycloak `typ` claim with the value `ID` is rejected. A token that says nothing about its kind is accepted, which is what most providers issue.

### 2. Nebraska Configuration

**Required:**
```bash
--oidc-client-id=your-public-client-id
--oidc-audience=your-nebraska-api-identifier
--oidc-issuer-url=https://your-oidc-provider.com
--oidc-admin-roles=nebraska-admin
--oidc-viewer-roles=nebraska-viewer
```

**Optional:**
```bash
--oidc-roles-path=roles                    # JSON path for roles (default: "roles")
--oidc-scopes=openid,profile,email         # OIDC scopes (default: "openid,profile,email")
--oidc-management-url=https://your-idp.com # Account management URL
--oidc-logout-url=https://your-idp.com/logout # Fallback logout URL
--oidc-skip-audience-check                 # Insecure migration escape hatch only
--ca-file=/path/to/ca.pem                 # Custom CA cert for TLS (see below)
```

### Custom CA Certificates

If your OIDC provider uses a certificate signed by a non-system CA (e.g., internal CA, Let's Encrypt staging), you can provide a custom CA certificate file:

```bash
--ca-file=/path/to/ca-bundle.pem
```

- The file should contain one or more PEM-encoded CA certificates
- Custom CAs are **added** to the system trust store (system-trusted CAs remain trusted)
- Applies to the OIDC HTTP client (discovery, JWKS, UserInfo) and the syncer HTTP client
- If the file is unreadable or contains no valid PEM certificates, Nebraska fails at startup with a clear error

### 3. Verification

1. Access Nebraska → redirects to OIDC provider
2. Authenticate → redirects back to Nebraska
3. Verify role-based access (admin vs viewer)
4. Confirm Nebraska started without an audience error in its logs
5. Confirm the access token carries the configured audience. If it does not, every API request returns 401 even though login appears to succeed

### 4. Token Expiration Recommendations

For optimal user experience:
- **Access Token Expiration**: Configure to 1-4 hours (industry standard) instead of default 5-15 minutes
- **SSO Session Duration**: Configure based on your security requirements:
  - **Idle Timeout**: 8-12 hours (user is logged out after this period of inactivity)
  - **Maximum Lifetime**: 1-7 days (user must re-authenticate after this period regardless of activity)
- **Session Configuration**:
  - **Keycloak**: Configure "SSO Session Max" and "SSO Session Idle Timeout" under Realm Settings → Sessions
  - **Auth0**: Configure "Maximum Session Lifetime" and "Idle Session Lifetime" under Tenant Settings → Advanced → Session Expiration
- **Note**: Since tokens are stored in-memory, when they expire after a page refresh, the OIDC provider automatically re-authenticates users if the SSO session is still active (no password re-entry required)

### 5. Troubleshooting

Visit the updated Nebraska documentation at `https://www.flatcar.org/docs/latest` 

| Issue | Solution |
|-------|----------|
| CORS errors | Enable CORS for Nebraska domain in OIDC provider |
| Invalid redirect URI | Add `https://your-domain.com/auth/callback` to allowed redirect URIs |
| Token validation failed | Check roles configuration and token claims |
| Access token audience rejected | Ensure the provider puts the configured `--oidc-audience` value in the access token's `aud` claim |
| "an ID token was presented as an access token" | The frontend must send the access token, not the ID token, as the Bearer credential |
| Nebraska refuses to start: "no access token audience configured" | Set `--oidc-audience`; to upgrade without reconfiguring the provider first, set `--oidc-skip-audience-check` and revisit it |
| User has no access | Verify user roles match configured roles |
| Frequent re-authentication | Increase access token expiration time in OIDC provider |
| JWT decode error (Auth0) | Ensure audience is set and Implicit grant is disabled |
| x509: certificate signed by unknown authority | Use `--ca-file` to trust your CA certificate |

**Debug JWT claims:**
```bash
# Decode JWT payload
echo "token" | cut -d. -f2 | tr '_-' '/+' | base64 -d | jq .
```

## References

- [ADR-001: OIDC Implementation](./architecture-decisions.md#adr-001-oidc-implementation-refactor)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
