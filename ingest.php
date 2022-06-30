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
$endpoint = $body["endpoint"];
$data = $body["data"];

$method = $endpoint["method"];

$dataSize = count($data);


$config = \Kafka\ProducerConfig::getInstance();
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList('127.0.0.1:9092');
$config->setBrokerVersion('1.0.0');
$config->setRequiredAck(1);
$config->setIsAsyn(false);
$config->setProduceInterval(500);
$producer = new \Kafka\Producer();



for($i = 0; $i < $dataSize; $i++) {
    $producer->send([
        [
            'topic' => 'postback', // Topic name already created using apache kafka's console commands.
            'value' => 'test1....message.', // TODO: Build correct info
            'key' => '',
        ],
    ]);
}

echo $method; // Testing that it works
