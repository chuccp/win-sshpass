#!/bin/bash
# Test all commands from README.md against Docker SSH server
set -e

BIN="./win-sshpass.exe"
HOST="localhost"
PORT="2222"
PASS="testpass"
USER="testuser"

# Use MSYS_NO_PATHCONV to prevent path mangling
export MSYS_NO_PATHCONV=1

# Use Windows-style paths for the binary
KEY_FILE="C:/Users/cao/AppData/Local/Temp/sshpass-test/keys/id_ed25519"
TEST_DIR="C:/Users/cao/AppData/Local/Temp/sshpass-test"

PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0
TOTAL=0

pass() { echo "  ✅ PASS: $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo "  ❌ FAIL: $1 — $2"; FAIL_COUNT=$((FAIL_COUNT + 1)); }
skip() { echo "  ⏭️  SKIP: $1 — $2"; SKIP_COUNT=$((SKIP_COUNT + 1)); }

run_test() {
    local desc="$1"
    local cmd="$2"
    local expected_ok="${3:-true}"
    TOTAL=$((TOTAL + 1))
    echo ""
    echo "── Test $TOTAL: $desc ──"
    echo "  CMD: $cmd"

    set +e
    output=$(eval "$cmd" 2>&1)
    rc=$?
    set -e

    echo "  RC: $rc"
    echo "  OUTPUT: ${output:0:300}"

    if [ "$expected_ok" = "true" ] && [ $rc -eq 0 ]; then
        pass "$desc"
    elif [ "$expected_ok" = "false" ] && [ $rc -ne 0 ]; then
        pass "$desc"
    else
        fail "$desc" "exit code $rc, expected_ok=$expected_ok"
    fi
}

echo "╔══════════════════════════════════════════════════════════╗"
echo "║   Testing all README.md commands - Docker SSH Server    ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
echo "Server: $USER@$HOST:$PORT (password: $PASS)"
echo "Binary: $BIN"

# Pre-test: Create test files
echo ""
echo "── Preparing test files ──"
mkdir -p "$TEST_DIR/local_test"
echo "test content 123" > "$TEST_DIR/local_test/file1.txt"
echo "test content 456" > "$TEST_DIR/local_test/file2.txt"
echo "test content 789" > "$TEST_DIR/local_test/file3.txt"
mkdir -p "$TEST_DIR/local_test/subdir"
echo "nested content" > "$TEST_DIR/local_test/subdir/nested.txt"
echo "password content" > "$TEST_DIR/pass.txt"
echo "$PASS" > "$TEST_DIR/pass_only.txt"

cat > "$TEST_DIR/server.config" << EOFCFG
host: $HOST
username: $USER
password: $PASS
port: $PORT
EOFCFG

echo "Test files ready"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 1: Quick Start"
echo "═══════════════════════════════════════════"

run_test "Password login + execute command (ssh style)" \
    "$BIN -p '$PASS' ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Password login with -u -h -P -c flags" \
    "$BIN -p '$PASS' -h $HOST -u $USER -P $PORT -c 'whoami'"

run_test "Private key login (-i flag)" \
    "$BIN -i '$KEY_FILE' ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Upload file (-local -remote)" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -local '$TEST_DIR/local_test/file1.txt' -remote //tmp/uploaded_file1.txt"

run_test "Download file (-d -remote -local)" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -d -remote //tmp/uploaded_file1.txt -local '$TEST_DIR/downloaded_file1.txt'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 2: SSH Login Variants"
echo "═══════════════════════════════════════════"

run_test "Password from environment variable (-e flag)" \
    "SSHPASS='$PASS' $BIN -e ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Password from file (-f flag)" \
    "$BIN -f '$TEST_DIR/pass_only.txt' ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Configuration file (-f with config)" \
    "$BIN -f '$TEST_DIR/server.config' -c 'whoami'"

run_test "Config file with positional command" \
    "$BIN -f '$TEST_DIR/server.config' 'whoami'"

run_test "SSH with -o StrictHostKeyChecking=no" \
    "$BIN -p '$PASS' ssh -o StrictHostKeyChecking=no -p $PORT $USER@$HOST 'whoami'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 3: File Transfer (SFTP)"
echo "═══════════════════════════════════════════"

run_test "Upload single file" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -local '$TEST_DIR/local_test/file1.txt' -remote //tmp/sftp_upload1.txt"

run_test "Upload multiple files (comma-separated)" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -local '$TEST_DIR/local_test/file1.txt,$TEST_DIR/local_test/file2.txt' -remote //tmp/sftp_multi/"

run_test "Upload directory (recursive)" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -local '$TEST_DIR/local_test/subdir' -remote //tmp/sftp_dir/"

run_test "Download single file" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -d -remote //tmp/sftp_upload1.txt -local '$TEST_DIR/sftp_downloaded.txt'"

run_test "Download directory" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -d -remote //tmp/sftp_dir -local '$TEST_DIR/sftp_downloaded_dir'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 4: SCP Style"
echo "═══════════════════════════════════════════"

run_test "SCP upload file" \
    "$BIN -p '$PASS' -P $PORT scp '$TEST_DIR/local_test/file1.txt' $USER@$HOST://tmp/scp_upload.txt"

run_test "SCP upload with capital -P port" \
    "$BIN -p '$PASS' scp -P $PORT '$TEST_DIR/local_test/file2.txt' $USER@$HOST://tmp/scp_upload2.txt"

run_test "SCP upload directory (-r flag)" \
    "$BIN -p '$PASS' -P $PORT scp -r '$TEST_DIR/local_test/subdir' $USER@$HOST://tmp/scp_dir/"

run_test "SCP download file" \
    "$BIN -p '$PASS' -P $PORT scp $USER@$HOST://tmp/scp_upload.txt '$TEST_DIR/scp_downloaded.txt'"

run_test "SCP download directory" \
    "$BIN -p '$PASS' -P $PORT scp -r $USER@$HOST://tmp/scp_dir '$TEST_DIR/scp_downloaded_dir'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 5: Rsync Style"
echo "═══════════════════════════════════════════"

run_test "Rsync upload" \
    "$BIN -p '$PASS' -P $PORT rsync -avz '$TEST_DIR/local_test/subdir' $USER@$HOST://tmp/rsync_upload/"

run_test "Rsync download" \
    "$BIN -p '$PASS' -P $PORT rsync -avz $USER@$HOST://tmp/rsync_upload '$TEST_DIR/rsync_downloaded'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 6: Complete Examples (README)"
echo "═══════════════════════════════════════════"

run_test "Example 1: Password login + command" \
    "$BIN -p '$PASS' ssh -p $PORT $USER@$HOST 'ls -la /tmp/testdata'"

run_test "Example 2: Key login + command" \
    "$BIN -i '$KEY_FILE' ssh -p $PORT $USER@$HOST 'ls -la /tmp/testdata'"

run_test "Example 3: Upload directory" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -local '$TEST_DIR/local_test' -remote //tmp/example_upload/"

run_test "Example 4: Download directory" \
    "$BIN -h $HOST -p '$PASS' -P $PORT -u $USER -d -remote //tmp/example_upload -local '$TEST_DIR/example_download'"

run_test "Example 5: SCP upload" \
    "$BIN -p '$PASS' -P $PORT scp '$TEST_DIR/local_test/file1.txt' $USER@$HOST://tmp/example_scp.txt"

run_test "Example 6: Env var password" \
    "SSHPASS='$PASS' $BIN -e ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Example 7: Operation timeout (-t 30)" \
    "$BIN -p '$PASS' -t 30 ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Example 8: Config file with positional command" \
    "$BIN -f '$TEST_DIR/server.config' 'whoami'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 7: Additional Parameters"
echo "═══════════════════════════════════════════"

run_test "Connection timeout (-ct 5)" \
    "$BIN -p '$PASS' -ct 5 ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Retry count (-retry 3)" \
    "$BIN -p '$PASS' -retry 3 ssh -p $PORT $USER@$HOST 'whoami'"

run_test "Version flag (-v)" \
    "$BIN -v"

run_test "Help flag (-help)" \
    "$BIN -help"

run_test "SSH with -c flag for command" \
    "$BIN -p '$PASS' -h $HOST -u $USER -P $PORT -c 'echo hello_world'"

# ═══════════════════════════════════════════
echo ""
echo "═══════════════════════════════════════════"
echo "  SECTION 8: Verification"
echo "═══════════════════════════════════════════"

run_test "Verify uploaded file exists on remote" \
    "$BIN -p '$PASS' ssh -p $PORT $USER@$HOST 'cat /tmp/uploaded_file1.txt'"

run_test "Verify downloaded file content" \
    "cat '$TEST_DIR/downloaded_file1.txt'"

run_test "Verify SCP uploaded file" \
    "$BIN -p '$PASS' ssh -p $PORT $USER@$HOST 'cat /tmp/scp_upload.txt'"

run_test "Verify rsync uploaded files" \
    "$BIN -p '$PASS' ssh -p $PORT $USER@$HOST 'ls /tmp/rsync_upload/'"

# ═══════════════════════════════════════════
echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║                      TEST SUMMARY                       ║"
echo "╠══════════════════════════════════════════════════════════╣"
printf "║  Total:  %-3s                                            ║\n" "$TOTAL"
printf "║  Passed: %-3s                                            ║\n" "$PASS_COUNT"
printf "║  Failed: %-3s                                            ║\n" "$FAIL_COUNT"
printf "║  Skipped: %-3s                                           ║\n" "$SKIP_COUNT"
echo "╚══════════════════════════════════════════════════════════╝"

if [ $FAIL_COUNT -gt 0 ]; then
    echo ""
    echo "⚠️  Some tests FAILED. See details above."
    exit 1
else
    echo ""
    echo "🎉 All tests PASSED!"
    exit 0
fi
