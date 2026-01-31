-- Migration: Drop upload_requests table

DROP TRIGGER IF EXISTS trigger_upload_requests_updated_at ON upload_requests;
DROP FUNCTION IF EXISTS update_upload_requests_updated_at();
DROP TABLE IF EXISTS upload_requests;
