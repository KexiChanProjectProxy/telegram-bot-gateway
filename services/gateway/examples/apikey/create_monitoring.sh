#!/bin/bash
# Example: Create read-only monitoring API key

set -e

APIKEY_BIN="./bin/apikey"
SERVICE_NAME="Monitoring Service"

echo "Creating read-only monitoring API key..."
echo "========================================="
echo ""

# Create the API key with high rate limit for monitoring
echo "1. Creating API key..."
OUTPUT=$($APIKEY_BIN create \
  --name "$SERVICE_NAME" \
  --description "Read-only access for monitoring" \
  --rate-limit 10000)

echo "$OUTPUT"
echo ""

# Extract API key ID
APIKEY_ID=$(echo "$OUTPUT" | grep "API Key ID:" | awk '{print $4}')

if [ -z "$APIKEY_ID" ]; then
  echo "Error: Failed to extract API key ID"
  exit 1
fi

# Grant read-only access to multiple chats
echo "2. Granting read-only access to multiple chats..."
for CHAT_ID in 5 8 10; do
  echo "   - Chat $CHAT_ID"
  $APIKEY_BIN grant-chat $APIKEY_ID $CHAT_ID --read
done
echo ""

# Show final permissions
echo "3. Final permissions:"
echo "-------------------"
$APIKEY_BIN show-permissions $APIKEY_ID
echo ""

echo "✓ Monitoring API key ready!"
echo ""
echo "This API key can:"
echo "  - Read messages from chats 5, 8, and 10"
echo "  - NO send permissions (read-only)"
echo "  - Use all bots (no restrictions)"
echo "  - Receive feedback from all chats (no restrictions)"
