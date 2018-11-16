<?php

class SocketLog {
    const SOCKET = '/var/www/logger.socket';

    public static function Debug($message, $context = array()){
        self::logEntry('Debug', $message, $context);
    }

    public static function Info($message, $context = array()){
        self::logEntry('Info', $message, $context);
    }

    public static function Notice($message, $context = array()){
        self::logEntry('Notice', $message, $context);
    }

    public static function Warning($message, $context = array()){
        self::logEntry('Warning', $message, $context);
    }

    public static function Error($message, $context = array()){
        self::logEntry('Error', $message, $context);
    }

    public static function Critical($message, $context = array()){
        self::logEntry('Critical', $message, $context);
    }

    public static function Alert($message, $context = array()){
        self::logEntry('Alert', $message, $context);
    }

    public static function Emergancy($message, $context = array()){
        self::logEntry('Emergancy', $message, $context);
    }

    protected static function logEntry($severity, $message, $context) {
        $object = array(
            'severity'  => $severity,
            'message'   => $message,
            'context'   => $context
        );

        $data = json_encode($object);

        if (strlen($data) > (96*1024)){
            $object = array(
                'severity'  => $severity,
                'message'   => $message,
                'context'   => 'Context removed as too large: ' . strlen($data) . ' bytes'
            );

            $data = json_encode($object);
        }

        self::logEntryEncoded($data);
    }

    protected static function logEntryEncoded($encodedMessage)
    {
        if (!file_exists(self::SOCKET)){
            \Log::error('SocketLogger: Socket Doesn\'t Exist!');
            return;
        }

        $errno = 0; $errstr = '';
        $fp = @stream_socket_client('unix://' . self::SOCKET, $errno, $errstr);
        if (!$fp){
            \Log::error('SocketLogger: Failed to open socket:' . $errno . ': ' . $errstr);
            return;
        }

        fwrite($fp, $encodedMessage);

        fclose($fp);
    }
}

// \SocketLog::Info('Test Message', array('test' => array('data' => 1)));
