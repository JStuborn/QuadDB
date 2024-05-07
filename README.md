# QuadDB

*Please note that this project is still under development and may not be suitable for production use. Contributions and feedback are welcome!*


This project is an **experimental** database management system (DBMS) developed in Go (Golang), tailored to proficiently manage NoSQL queries. 
It offers fundamental features including:
- Default AES encryption for databases with customizable keys.
- Full configurability across all parameters.
- Support for storing and querying JSON documents in a document-oriented database format.
- Built-in GZ compression functionality for optimized storage.


## Planned Functionalities

1. **Authentication and Authorization**:
   - Implement user authentication and authorization mechanisms to control access to the database.
   - Allow users to create accounts, log in, and manage their permissions for accessing databases and performing CRUD operations.

2. **Indexing**:
   - Add support for indexing specific fields within JSON documents to improve query performance.
   - Implement indexing mechanisms such as B-tree or hash indexes to enable faster data retrieval.

3. **Backup and Restore**:
   - Provide functionality to create backups of databases and restore them when needed.
   - Support automated backup schedules and retention policies for managing backup files.

4. **Replication and Sharding**:
   - Implement database replication to maintain redundant copies of data for fault tolerance and high availability.
   - Introduce sharding capabilities to distribute data across multiple servers for horizontal scalability.

5. **Query Language Enhancements**:
   - Extend the query language to support more complex operations such as joins, aggregations, and transactions.
   - Improve query optimization techniques to enhance overall performance for complex queries.

6. **Monitoring and Logging**:
   - Integrate monitoring tools to track database performance metrics such as CPU usage, memory consumption, and disk I/O.
   - Enable logging functionality to record database events, errors, and user activities for auditing purposes.

7. **Data Validation and Schema Enforcement**:
   - Implement data validation rules to ensure that incoming JSON documents adhere to predefined schemas.
   - Enforce schema constraints to maintain data integrity and prevent invalid data from being stored in the database.

8. **Clustering and Load Balancing**:
   - Introduce clustering support to create clusters of database nodes for improved scalability and fault tolerance.
   - Implement load balancing mechanisms to evenly distribute client requests across multiple nodes within the cluster.

9. **Dockerization and Orchestration**:
   - Dockerize the QuadDB application to simplify deployment and management using containerization.
   - Provide support for container orchestration platforms such as Kubernetes for automating deployment, scaling, and resource management.

10. **Advanced Encryption Options**:
    - Extend encryption capabilities to support additional encryption algorithms besides AES.
    - Allow users to choose encryption key management options such as key rotation and key vault integration for enhanced security.

