<?php
require 'vendor/autoload.php';
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
$data = $body["data"];

$method = $body["endpoint"]["method"];

$config = \Kafka\ProducerConfig::getInstance();
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList('127.0.0.1:9092');
$config->setBrokerVersion('1.0.0');
$config->setRequiredAck(1);
$config->setIsAsyn(false);
$config->setProduceInterval(500);
$producer = new \Kafka\Producer();

for($i = 0; $i < count($data); $i++) { // Send postback for each element on the data array
    $payload = $body;
    $payload["data"] = $data[$i];
    $payload = json_encode($payload);

    $producer->send([
        [
            'topic' => 'postback', // Topic name already created using apache kafka's console commands.
            'value' => $payload,
            'key' => '',
        ],
    ]);
}

echo $method; // Testing that it works
