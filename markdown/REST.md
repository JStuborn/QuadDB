# RESTish-API Documentation

## Overview
This document provides an overview of the available API routes for interacting with QuadDB.

## Endpoints

### Ping
#### `GET /ping`
**Description**: Check the health status of the API.
**Response**: {`message`: `pong`}

### Get Documents by Database
#### `GET /api/v1/docs/:db`
**Description**: Retrieve documents from a specific database.
**Parameters**:
- `db`: Database name (path parameter)
- `page`: Page number (query parameter, default: 1)
- `size`: Number of records per page (query parameter, default: 5)
**Response**: {`_resp`: `time_taken`, `_num`: number_of_documents, `documents`: [ { `id`: `document_id`, `data`: `document_data` } ]}

### Create Documents
#### `POST /api/v1/docs/:db`
**Description**: Create documents in a specific database.
**Parameters**:
- `db`: Database name (path parameter)
**Request Body**: 
{
  `id`: `document_id`, 
  `data`: `document_data`
}
**Response**: {`_resp`: `time_taken`, `message`: `Documents created successfully`}

### Search Documents by Field Value
#### `GET /api/v1/docs/:db/search`
**Description**: Search documents in a specific database by a field value.
**Parameters**:
- `db`: Database name (path parameter)
- `field`: Field name to search (query parameter)
- `value`: Value to search for (query parameter)
**Response**: {`_resp`: `time_taken`, `_num`: number_of_documents, `documents`: [ { `id`: `document_id`, `data`: `document_data` } ]}

### Get Document by Key
#### `GET /api/v1/docs/:db/:key`
**Description**: Retrieve a document by its key from a specific database.
**Parameters**:
- `db`: Database name (path parameter)
- `key`: Document key (path parameter)
**Response**: {`_resp`: `time_taken`, `data`: `document_data`}

### Update Document by Key
#### `PUT /api/v1/docs/:db/:key`
**Description**: Update a document by its key in a specific database.
**Parameters**:
- `db`: Database name (path parameter)
- `key`: Document key (path parameter)
**Request Body**: document_data
**Response**: {`_resp`: `time_taken`, `message`: `Document updated successfully`}

### Delete Document by Key
#### `DELETE /api/v1/docs/:db/:key`
**Description**: Delete a document by its key from a specific database.
**Parameters**:
- `db`: Database name (path parameter)
- `key`: Document key (path parameter)
**Response**: {`_resp`: `time_taken`, `message`: `Document deleted successfully`}

### Get Admin Information
#### `GET /api/v1/docs/updates`
**Description**: Retrieve information about the last used database, last update time, last added record, and last read record.
**Response**: 
{
  `last_used_db`: `last_used_db`,
  `last_update_time`: `last_update_time`,
  `last_added_record`: `last_added_record`,
  `last_read_record`: `last_read_record`
}

### Get Collections
#### `GET /api/v1/docs/collections`
**Description**: Retrieve the count of documents in each database collection.
**Response**: 
{
  `database_name`: document_count
}
