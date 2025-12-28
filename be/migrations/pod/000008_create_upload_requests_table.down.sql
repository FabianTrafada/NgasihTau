-- Migration: Drop upload_requests table
-- Requirements: 4.1, 4.3, 4.6

DROP TRIGGER IF EXISTS trigger_upload_requests_updated_at ON upload_requests;
DROP FUNCTION IF EXISTS update_upload_requests_updated_at();
DROP TABLE IF EXISTS upload_requests;
