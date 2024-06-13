<div align="center" style="margin-bottom: 15px;">
  <img src="https://github.com/CyberDefenseEd/QuadDB/raw/main/images/icon.svg" width="120" alt="QuadDB Logo" style="background-color: black;">
</div>

---

# Quadrium Database

*Please note that this project is still under development and may not be suitable for production use. Contributions and feedback are welcome!*

Quadrium (a.k.a QuadDB) is an experimental database management system (DBMS) developed in Go (Golang), tailored to proficiently manage NoSQL queries.

## Current Features

- **Default AES Encryption**: QuadDB offers default AES encryption for database collections, ensuring data security.
- **Configurability**: Full configurability across all parameters, including server port, AES password key, and data storage directory.
- **Document-Oriented Format**: Support for storing and querying [msgpack](https://github.com/vmihailenco/msgpack) documents in a document-oriented database format.
- **GZ Compression**: Built-in GZ compression functionality for optimized storage.
- **Admin Dashboard**: A simple admin dashboard for viewing record counts and collections.
- **UID Randomization**: Simple and blazing fast UUID4 generation for document key:value pairs.

## What is the QDB extention?
The .qdb extension is used for files that store JSONL (JSON Lines) data, where each line is a separate JSON object. To ensure data security and efficient storage, these files undergo two key processes:

1. **AES Encryption**
   - The JSONL data is encrypted using the Advanced Encryption Standard (AES), a widely recognized encryption standard that provides robust security for sensitive information. This ensures that the data is protected from unauthorized access and tampering.
2. **GZip Compression**
   - After encryption, the data is compressed using the GZip compression method. GZip is a popular compression algorithm that reduces the file size, making it easier to store and transfer while maintaining the integrity of the original data.

Together, these processes provide a secure and efficient way to store JSONL data, making the .qdb extension suitable for applications that require both data protection and optimization of storage space.

## Planned Functionalities & Rest API
Both have been moved to our wiki [here](https://github.com/CyberDefenseEd/QuadDB/wiki)
