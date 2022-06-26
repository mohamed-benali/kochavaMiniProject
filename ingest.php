<?php

/* POST Request example
 {
	"endpoint":{
		"method":"GET",
		"url":"http://sample_domain_endpoint.com/data?title={mascot}&image={location}&foo={bar}"
	},
	"data":[
		{
		"mascot":"Gopher",
		"location":"https://blog.golang.org/gopher/gopher.png"
		}
	]
}
 */

$body = json_decode(file_get_contents('php://input'), true); // Get body as a JSON
$endpoint = $body["endpoint"];
$data = $body["data"];

$method = $endpoint["method"];

echo $method; // Testing that it works
