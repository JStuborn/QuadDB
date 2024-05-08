<div align="center" style="margin-bottom: 15px;">
  <img src="https://github.com/CyberDefenseEd/QuadDB/raw/main/images/icon.svg" width="120" alt="QuadDB Logo" style="background-color: black;">
</div>

---

# Quadrium Database

*Please note that this project is still under development and may not be suitable for production use. Contributions and feedback are welcome!*

Quadrium (a.k.a QuadDB) is an experimental database management system (DBMS) developed in Go (Golang), tailored to proficiently manage NoSQL queries.

## Features

- **Default AES Encryption**: QuadDB offers default AES encryption for database collections, ensuring data security.
- **Configurability**: Full configurability across all parameters, including server port, AES password key, and data storage directory.
- **Document-Oriented Format**: Support for storing and querying JSON documents in a document-oriented database format.
- **GZ Compression**: Built-in GZ compression functionality for optimized storage.

## Planned Functionalities

1. **Authentication and Authorization**:
   - Implement user authentication and authorization mechanisms.
   - Allow users to create accounts, log in, and manage permissions.

2. **Indexing**:
   - Add support for indexing specific fields within JSON documents.
   - Implement indexing mechanisms such as B-tree or hash indexes.

3. **Backup and Restore**:
   - Provide functionality to create backups of databases and restore them.
   - Support automated backup schedules and retention policies.

4. **Replication and Sharding**:
   - Implement database replication for fault tolerance and high availability.
   - Introduce sharding capabilities for horizontal scalability.

5. **Query Language Enhancements**:
   - Extend the query language to support more complex operations.
   - Improve query optimization techniques.

6. **Monitoring and Logging**:
   - Integrate monitoring tools to track database performance metrics.
   - Enable logging functionality for auditing purposes.

7. **Data Validation and Schema Enforcement**:
   - Implement data validation rules to ensure data integrity.
   - Enforce schema constraints for data validation.

8. **Clustering and Load Balancing**:
   - Introduce clustering support for improved scalability.
   - Implement load balancing mechanisms for distributing client requests.

9. **Dockerization and Orchestration**:
   - Dockerize the QuadDB application for simplified deployment.
   - Provide support for container orchestration platforms such as Kubernetes.

10. **Advanced Encryption Options**:
    - Extend encryption capabilities to support any string converted to SHA256.
    - Allow users to choose encryption key management options for enhanced security.

## Get Started

To get started with QuadDB, visit the [GitHub repository](https://github.com/CyberDefenseEd/QuadDB) and follow the installation instructions.

Your feedback and contributions are valuable! Feel free to open issues or pull requests on GitHub.
