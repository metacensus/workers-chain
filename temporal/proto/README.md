# Table of Contents

- [aws.s3.v1](#aws-s3-v1)
  - Services
    - [aws.s3.v1.S3Provider](#aws-s3-v1-s3provider)
      - [Activities](#aws-s3-v1-s3provider-activities)
        - [aws.s3.v1.GetObject](#aws-s3-v1-getobject-activity)
        - [aws.s3.v1.PutObject](#aws-s3-v1-putobject-activity)
        - [aws.s3.v1.DeleteObject](#aws-s3-v1-deleteobject-activity)
  - Messages
    - [aws.s3.v1.S3DeleteOutput](#aws-s3-v1-s3deleteoutput)
    - [aws.s3.v1.S3Object](#aws-s3-v1-s3object)
    - [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)

<a name="aws-s3-v1"></a>
# aws.s3.v1

<a name="aws-s3-v1-services"></a>
## Services

<a name="aws-s3-v1-s3provider"></a>
## aws.s3.v1.S3Provider

<pre>
S3Provider service defines activities for interacting with AWS S3
</pre>   

<a name="aws-s3-v1-s3provider-activities"></a>
### Activities

---
<a name="aws-s3-v1-getobject-activity"></a>
### aws.s3.v1.GetObject

<pre>
GetObject retrieves an object from S3
</pre>

**Input:** [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table> 

---
<a name="aws-s3-v1-putobject-activity"></a>
### aws.s3.v1.PutObject

<pre>
PutObject writes an object to S3
</pre>

**Input:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table> 

---
<a name="aws-s3-v1-deleteobject-activity"></a>
### aws.s3.v1.DeleteObject

<pre>
DeleteObject deletes an object from S3
</pre>

**Input:** [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3DeleteOutput](#aws-s3-v1-s3deleteoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>delete_marker</td>
<td>bool</td>
<td><pre>
delete_marker indicates if a delete marker was created (for versioned buckets)<br>

json_name: deleteMarker
go_name: DeleteMarker</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>   

<a name="aws-s3-v1-messages"></a>
## Messages

<a name="aws-s3-v1-s3deleteoutput"></a>
### aws.s3.v1.S3DeleteOutput

<pre>
S3DeleteOutput represents the result of a delete operation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>delete_marker</td>
<td>bool</td>
<td><pre>
delete_marker indicates if a delete marker was created (for versioned buckets)<br>

json_name: deleteMarker
go_name: DeleteMarker</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>



<a name="aws-s3-v1-s3object"></a>
### aws.s3.v1.S3Object

<pre>
S3Object represents a complete S3 object with metadata
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>



<a name="aws-s3-v1-s3objectref"></a>
### aws.s3.v1.S3ObjectRef

<pre>
S3ObjectRef references an S3 object without its data
Used for operations that only need bucket and key (read, delete)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>

