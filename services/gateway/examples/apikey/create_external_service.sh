#!/bin/bash
# Example: Create API key for external service with granular permissions

set -e

APIKEY_BIN="./bin/apikey"
SERVICE_NAME="External Service Demo"
CHAT_ID=5
BOT_ID=2

echo "Creating API key for external service..."
echo "========================================="
echo ""

# Create the API key
echo "1. Creating API key..."
OUTPUT=$($APIKEY_BIN create \
  --name "$SERVICE_NAME" \
  --description "Example: External service with limited permissions" \
  --rate-limit 5000 \
  --expires 1y)

echo "$OUTPUT"
echo ""

# Extract API key ID from output
APIKEY_ID=$(echo "$OUTPUT" | grep "API Key ID:" | awk '{print $4}')

if [ -z "$APIKEY_ID" ]; then
  echo "Error: Failed to extract API key ID"
  exit 1
fi

echo "API Key ID: $APIKEY_ID"
echo ""

# Grant chat permissions
echo "2. Granting chat permissions..."
$APIKEY_BIN grant-chat $APIKEY_ID $CHAT_ID --read --send
echo ""

# Restrict to specific bot
echo "3. Restricting to bot $BOT_ID..."
$APIKEY_BIN grant-bot $APIKEY_ID $BOT_ID
echo ""

# Allow feedback from the same chat
echo "4. Allowing feedback from chat $CHAT_ID..."
$APIKEY_BIN grant-feedback $APIKEY_ID $CHAT_ID
echo ""

# Show final permissions
echo "5. Final permissions:"
echo "-------------------"
$APIKEY_BIN show-permissions $APIKEY_ID
echo ""

echo "✓ Setup complete!"
echo ""
echo "This API key can now:"
echo "  - Read and send messages to chat $CHAT_ID"
echo "  - ONLY use bot $BOT_ID (restricted)"
echo "  - ONLY receive feedback from chat $CHAT_ID (restricted)"
