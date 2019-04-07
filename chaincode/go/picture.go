/*
 SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke pictures ====
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["initPicture","picture1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["initPicture","picture2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["initPicture","picture3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["transferPicture","picture2","jerry"]}'
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["transferPicturesBasedOnGeneration","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n pictures -c '{"Args":["delete","picture1"]}'

// ==== Query pictures ====
// peer chaincode query -C myc1 -n pictures -c '{"Args":["readPicture","picture1"]}'
// peer chaincode query -C myc1 -n pictures -c '{"Args":["getPicturesByRange","picture1","picture3"]}'
// peer chaincode query -C myc1 -n pictures -c '{"Args":["getHistoryForPicture","picture1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n pictures -c '{"Args":["queryPicturesByOwner","tom"]}'
// peer chaincode query -C myc1 -n pictures -c '{"Args":["queryPictures","{\"selector\":{\"owner\":\"tom\"}}"]}'

// Rich Query with Pagination (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n pictures -c '{"Args":["queryPicturesWithPagination","{\"selector\":{\"owner\":\"tom\"}}","3",""]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This pictures02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support pictures02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_pictures/_index
//

// Index for docType, owner, size (descending order).
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_pictures/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n pictures -c '{"Args":["queryPictures","{\"selector\":{\"docType\":\"picture\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n pictures -c '{"Args":["queryPictures","{\"selector\":{\"docType\":{\"$eq\":\"picture\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type picture struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Generation      string `json:"generation"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initPicture" { //create a new picture
		return t.initPicture(stub, args)
	} else if function == "transferPicture" { //change owner of a specific picture
		return t.transferPicture(stub, args)
	} else if function == "transferPicturesBasedOnGeneration" { //transfer all pictures of a certain generation
		return t.transferPicturesBasedOnGeneration(stub, args)
	} else if function == "delete" { //delete a picture
		return t.delete(stub, args)
	} else if function == "readPicture" { //read a picture
		return t.readPicture(stub, args)
	} else if function == "queryPicturesByOwner" { //find pictures for owner X using rich query
		return t.queryPicturesByOwner(stub, args)
	} else if function == "queryPictures" { //find pictures based on an ad hoc rich query
		return t.queryPictures(stub, args)
	} else if function == "getHistoryForPicture" { //get history of values for a picture
		return t.getHistoryForPicture(stub, args)
	} else if function == "getPicturesByRange" { //get pictures based on range query
		return t.getPicturesByRange(stub, args)
	} else if function == "getPicturesByRangeWithPagination" {
		return t.getPicturesByRangeWithPagination(stub, args)
	} else if function == "queryPicturesWithPagination" {
		return t.queryPicturesWithPagination(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initPicture - create a new picture, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initPicture(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init picture")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	pictureName := args[0]
	generation := strings.ToLower(args[1])
	owner := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if picture already exists ====
	pictureAsBytes, err := stub.GetState(pictureName)
	if err != nil {
		return shim.Error("Failed to get picture: " + err.Error())
	} else if pictureAsBytes != nil {
		fmt.Println("This picture already exists: " + pictureName)
		return shim.Error("This picture already exists: " + pictureName)
	}

	// ==== Create picture object and marshal to JSON ====
	objectType := "picture"
	picture := &picture{objectType, pictureName, generation, size, owner}
	pictureJSONasBytes, err := json.Marshal(picture)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the picture json string manually if you don't want to use struct marshalling
	//pictureJSONasString := `{"docType":"Picture",  "name": "` + pictureName + `", "generation": "` + generation + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//pictureJSONasBytes := []byte(str)

	// === Save picture to state ===
	err = stub.PutState(pictureName, pictureJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the picture to enable generation-based range queries, e.g. return all blue pictures ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~generation~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~generation~*
	indexName := "generation~name"
	generationNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{picture.Generation, picture.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the picture.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(generationNameIndexKey, value)

	// ==== Picture saved and indexed. Return success ====
	fmt.Println("- end init picture")
	return shim.Success(nil)
}

// ===============================================
// readPicture - read a picture from chaincode state
// ===============================================
func (t *SimpleChaincode) readPicture(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the picture to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the picture from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Picture does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a picture key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var pictureJSON picture
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	pictureName := args[0]

	// to maintain the generation~name index, we need to read the picture first and get its generation
	valAsbytes, err := stub.GetState(pictureName) //get the picture from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + pictureName + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Picture does not exist: " + pictureName + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &pictureJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + pictureName + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(pictureName) //remove the picture from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "generation~name"
	generationNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{pictureJSON.Generation, pictureJSON.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(generationNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a picture by setting a new owner name on the picture
// ===========================================================
func (t *SimpleChaincode) transferPicture(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	pictureName := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferPicture ", pictureName, newOwner)

	pictureAsBytes, err := stub.GetState(pictureName)
	if err != nil {
		return shim.Error("Failed to get picture:" + err.Error())
	} else if pictureAsBytes == nil {
		return shim.Error("Picture does not exist")
	}

	pictureToTransfer := picture{}
	err = json.Unmarshal(pictureAsBytes, &pictureToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	pictureToTransfer.Owner = newOwner //change the owner

	pictureJSONasBytes, _ := json.Marshal(pictureToTransfer)
	err = stub.PutState(pictureName, pictureJSONasBytes) //rewrite the picture
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferPicture (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

// ===========================================================================================
// getPicturesByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getPicturesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- getPicturesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferPicturesBasedOnGeneration will transfer pictures of a given generation to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against generation~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferPicturesBasedOnGeneration(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "generation", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	generation := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferPicturesBasedOnGeneration ", generation, newOwner)

	// Query the generation~name index by generation
	// This will execute a key range query on all keys starting with 'generation'
	generationedPictureResultsIterator, err := stub.GetStateByPartialCompositeKey("generation~name", []string{generation})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer generationedPictureResultsIterator.Close()

	// Iterate through result set and for each picture found, transfer to newOwner
	var i int
	for i = 0; generationedPictureResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the picture name from the composite key
		responseRange, err := generationedPictureResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the generation and name from generation~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedGeneration := compositeKeyParts[0]
		returnedPictureName := compositeKeyParts[1]
		fmt.Printf("- found a picture from index:%s generation:%s name:%s\n", objectType, returnedGeneration, returnedPictureName)

		// Now call the transfer function for the found picture.
		// Re-use the same function that is used to transfer individual pictures
		response := t.transferPicture(stub, []string{returnedPictureName, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s pictures to %s", i, generation, newOwner)
	fmt.Println("- end transferPicturesBasedOnGeneration: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryPicturesByOwner queries for pictures based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryPicturesByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"picture\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryPictures uses a query string to perform a query for pictures.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryPicturesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryPictures(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ====== Pagination =========================================================================
// Pagination provides a method to retrieve records with a defined pagesize and
// start point (bookmark).  An empty string bookmark defines the first "page" of a query
// result.  Paginated queries return a bookmark that can be used in
// the next query to retrieve the next page of results.  Paginated queries extend
// rich queries and range queries to include a pagesize and bookmark.
//
// Two examples are provided in this example.  The first is getPicturesByRangeWithPagination
// which executes a paginated range query.
// The second example is a paginated query for rich ad-hoc queries.
// =========================================================================================

// ====== Example: Pagination with Range Query ===============================================
// getPicturesByRangeWithPagination performs a range query based on the start & end key,
// page size and a bookmark.

// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// ===========================================================================================
func (t *SimpleChaincode) getPicturesByRangeWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	startKey := args[0]
	endKey := args[1]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[3]

	resultsIterator, responseMetadata, err := stub.GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	fmt.Printf("- getPicturesByRange queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return shim.Success(buffer.Bytes())
}

// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryPicturesWithPagination uses a query string, page size and a bookmark to perform a query
// for pictures. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryPicturesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (t *SimpleChaincode) queryPicturesWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForPicture(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	pictureName := args[0]

	fmt.Printf("- start getHistoryForPicture: %s\n", pictureName)

	resultsIterator, err := stub.GetHistoryForKey(pictureName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the picture
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON picture)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForPicture returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
