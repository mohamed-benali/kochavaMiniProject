<?php
require 'vendor/autoload.php';

const BROKER_ADDRESS = '127.0.0.1:9092'; // Apache kafka address

$body = json_decode(file_get_contents('php://input'), true); // Get body as a JSON
$data = $body["data"];

$method = $body["endpoint"]["method"];

$config = \Kafka\ProducerConfig::getInstance(); // Connect with Kafka as producer
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList(BROKER_ADDRESS);
$config->setBrokerVersion('1.0.0');
$config->setRequiredAck(1);
$config->setIsAsyn(false);
$config->setProduceInterval(500);
$producer = new \Kafka\Producer();

// I assume there is at least one postback (count($data) >= 1)
for($i = 0; $i < count($data); $i++) { // Send postback for each element on the data array
    $payload = $body;
    $payload["data"] = $data[$i]; // Each postback is sent as one object
    $payload = json_encode($payload);

    $producer->send([
        [
            'topic' => 'postback', // Topic name already created using apache kafka's console commands.
            'value' => $payload,
            'key' => '',
        ],
    ]);
}

echo $method; // Writes the method
