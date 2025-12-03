# triple-s

## Overview of the Project

The tool allows users to manage storage buckets and objects through an HTTP-based REST API, providing features such as creating and listing buckets, uploading, retrieving, and deleting files, and storing associated metadata. By completing this project, I gained invaluable insights into the inner workings of cloud storage systems.

---

## Functional Highlights

### Bucket Management
- **Create Buckets:** Users can create buckets with S3-compliant naming conventions.
- **List Buckets:** List all available buckets with metadata in XML format.
- **Delete Buckets:** Safely remove empty buckets while handling conflicts.

### Object Operations
- **Upload Objects:** Store objects within specified buckets, updating metadata in real-time.
- **Retrieve Objects:** Fetch objects via a direct HTTP GET request, complete with headers for MIME types.
- **Delete Objects:** Securely delete objects and update the corresponding metadata.

### Directory Structure
- Buckets and objects are stored in an organized file system.
- Each bucket has a dedicated metadata file for object details.

---
### API Endpoints

#### Buckets

| Method | Endpoint          | Description                       | Response                   |
|--------|-------------------|-----------------------------------|----------------------------|
| PUT   | `/{BucketName}`         | Creates a new bucket.             | 201 Created                |
| GET    | `/`         | Retrieves all buckets.             | 200 OK                     |
| DELETE    | `/{BucketName}`    | Deletes a specific bucket. | 204 No content     |


#### Objects

| Method | Endpoint          | Description                        | Response                   |
|--------|-------------------|------------------------------------|----------------------------|
| PUT       | `/{BucketName}/{ObjectKey}`      | Adds a new object.               | 201 Created                |
| GET       | `/{BucketName}/{ObjectKey}`      | Retrieves a specific object binary data.          | 200 OK                     |
| DELETE    | `/{BucketName}/{ObjectKey}`      | Deletes a specific object.    | 204 No content     |


## Lessons Learned

- Gained in-depth knowledge of RESTful API design and HTTP protocol fundamentals.
- Enhanced my understanding of data persistence techniques, including file-based metadata management.
- Developed practical skills in debugging and error handling for scalable solutions.

---

## Usage Instructions

To run the program:

```bash
$ ./triple-s
```

For help:

```bash
$ ./triple-s --help
```

Start the server:
   ```bash
   $ ./triple-s --port 8080 --dir ./data/
   ```
   Here, `8080` is the port, and `./data/` is the directory for storing files.
