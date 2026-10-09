<?php

namespace Brickred\Exchange;

abstract class Binary64
{
    protected $high_;
    protected $low_;

    abstract public function getValue();

    public function __construct($str='')
    {
        $this->reset(0, 0);
        if ($str !== '') {
            $this->fromString($str);
        }
    }

    public function getHighInt32()
    {
        return $this->high_;
    }

    public function getLowInt32()
    {
        return $this->low_;
    }

    public function setHighInt32($high)
    {
        $this->high_ = Codec::convertToInt32($high);
    }

    public function setLowInt32($low)
    {
        $this->low_ = Codec::convertToInt32($low);
    }

    public function reset($high = 0, $low = 0)
    {
        $this->setHighInt32($high);
        $this->setLowInt32($low);
    }

    public function encode()
    {
        return Codec::writeInt32($this->high_).
               Codec::writeInt32($this->low_);
    }

    public function decodeFromStream($s)
    {
        $this->high_ = Codec::readInt32($s);
        $this->low_ = Codec::readInt32($s);
    }

    public function decode($buf)
    {
        $s = Codec::openStreamForBuffer($buf);
        $this->decodeFromStream($s);
        fclose($s);
    }

    public function toString()
    {
        return (string)($this->getValue());
    }

    public function fromString($str)
    {
        $str = bcmod($str, '18446744073709551616');
        if (bccomp($str, '0') < 0) {
            $str = bcadd($str, '18446744073709551616');
        }
        $this->setHighInt32(bcdiv($str, '4294967296'));
        $this->setLowInt32(bcmod($str, '4294967296'));
    }
}

final class Int64 extends Binary64
{
    public function getValue()
    {
        if (PHP_INT_SIZE === 8) {
            return $this->high_ << 32 | ($this->low_ & 0xffffffff);
        }

        if ($this->high_ === 0) {
            if ($this->low_ < 0) {
                return $this->low_ + 4294967296;
            } else {
                return $this->low_;
            }
        } else if ($this->high_ === -1) {
            if ($this->low_ < 0) {
                return $this->low_;
            } else {
                return $this->low_ - 4294967296;
            }
        }

        if ($this->low_ < 0) {
            $unsigned_low = $this->low_ + 4294967296;
        } else {
            $unsigned_low = $this->low_;
        }

        return bcadd(bcmul($this->high_, '4294967296'), $unsigned_low);
    }
}

final class UInt64 extends Binary64
{
    public function getValue()
    {
        if (PHP_INT_SIZE === 8) {
            $unsigned_high = $this->high_ & 0xffffffff;
            $unsigned_low = $this->low_ & 0xffffffff;
            if ($this->high_ > 0) {
                return $unsigned_high << 32 | $unsigned_low;
            } else {
                return bcadd(bcmul($unsigned_high, '4294967296'),
                             $unsigned_low);
            }
        }

        if ($this->low_ < 0) {
            $unsigned_low = $this->low_ + 4294967296;
        } else {
            $unsigned_low = $this->low_;
        }

        if ($this->high_ === 0) {
            return $unsigned_low;
        }

        if ($this->high_ < 0) {
            $unsigned_high = $this->high_ + 4294967296;
        } else {
            $unsigned_high = $this->high_;
        }

        return bcadd(bcmul($unsigned_high, '4294967296'), $unsigned_low);
    }
}

final class CodecException extends \Exception
{
}

final class Codec
{
    public static function openStreamForBuffer($buf)
    {
        $s = fopen('php://memory', 'r+');
        fwrite($s, $buf);
        rewind($s);
        return $s;
    }

    public static function convertToInt32($var)
    {
        if (is_string($var)) {
            $var = (float)$var;
        }

        if (is_float($var)) {
            $var = $var >= 0 ? floor($var) : ceil($var);
            $var = fmod($var, 4294967296.0);
            if ($var < 0) {
                $var += 4294967296.0;
            }
            if ($var > 2147483647.0) {
                $var -= 4294967296.0;
            }
            return (int)$var;
        }

        if (PHP_INT_SIZE === 8) {
            $var &= 0xffffffff;
            if ($var > 2147483647) {
                $var -= 4294967296;
            }
        }

        return (int)$var;
    }

    public static function readUInt8($s)
    {
        $ret = unpack('C', fread($s, 1));
        if ($ret === false) {
            throw new CodecException('read u8 failed');
        }
        return $ret[1];
    }

    public static function readUInt16($s)
    {
        $ret = unpack('n', fread($s, 2));
        if ($ret === false) {
            throw new CodecException('read u16 failed');
        }
        return $ret[1];
    }

    public static function readUInt32($s)
    {
        $ret = unpack('N', fread($s, 4));
        if ($ret === false) {
            throw new CodecException('read u32 failed');
        }

        $var = $ret[1];
        if ($var < 0) {
            return $var + 4294967296;
        }

        return $var;
    }

    public static function readUInt64($s)
    {
        $var = new UInt64();
        $var->decodeFromStream($s);

        return $var;
    }

    public static function readUInt16V($s)
    {
        $var = self::readUInt8($s);
        if ($var < 255) {
            return $var;
        } else {
            return self::readUInt16($s);
        }
    }

    public static function readUInt32V($s)
    {
        $var = self::readUInt8($s);
        if ($var < 254) {
            return $var;
        } else if ($var === 254) {
            return self::readUInt16($s);
        } else {
            return self::readUInt32($s);
        }
    }

    public static function readUInt64V($s)
    {
        $var = self::readUInt8($s);
        if ($var < 253) {
            return new UInt64($var);
        } else if ($var === 253) {
            return new UInt64(self::readUInt16($s));
        } else if ($var === 254) {
            return new UInt64(self::readUInt32($s));
        } else {
            return self::readUInt64($s);
        }
    }

    public static function readInt8($s)
    {
        $var = self::readUInt8($s);
        if ($var > 127) {
            $var -= 256;
        }

        return $var;
    }

    public static function readInt16($s)
    {
        $var = self::readUInt16($s);
        if ($var > 32767) {
            $var -= 65536;
        }

        return $var;
    }

    public static function readInt32($s)
    {
        $var = self::readUInt32($s);
        if ($var > 2147483647) {
            $var -= 4294967296;
        }

        return (int)$var;
    }

    public static function readInt64($s)
    {
        $var = new Int64();
        $var->decodeFromStream($s);

        return $var;
    }

    public static function readInt16V($s)
    {
        $var = self::readUInt16V($s);
        if ($var > 32767) {
            $var -= 65536;
        }

        return $var;
    }

    public static function readInt32V($s)
    {
        $var = self::readUInt32V($s);
        if ($var > 2147483647) {
            $var -= 4294967296;
        }

        return (int)$var;
    }

    public static function readInt64V($s)
    {
        $var = self::readUInt8($s);
        if ($var < 253) {
            return new Int64($var);
        } else if ($var === 253) {
            return new Int64(self::readUInt16($s));
        } else if ($var === 254) {
            return new Int64(self::readUInt32($s));
        } else {
            return self::readInt64($s);
        }
    }

    public static function readInt16VZ($s)
    {
        return self::zigzagDecode16(self::readUInt16V($s));
    }

    public static function readInt32VZ($s)
    {
        return self::zigzagDecode32(self::readUInt32V($s));
    }

    public static function readInt64VZ($s)
    {
        return self::zigzagDecode64(self::readUInt64V($s));
    }

    public static function readBool($s)
    {
        $var = self::readUInt8($s);
        if ($var === 0) {
            return false;
        } else {
            return true;
        }
    }

    public static function readLength($s)
    {
        return self::readUInt32V($s);
    }

    public static function readString($s)
    {
        $length = self::readLength($s);
        if ($length === 0) {
            return '';
        }

        $var = fread($s, $length);
        if (strlen($var) !== $length) {
            throw new CodecException('read string failed');
        }

        return $var;
    }

    public static function readStruct($s, $struct_name)
    {
        $var = new $struct_name();
        $var->decodeFromStream($s);

        return $var;
    }

    public static function readList($s, $read_func)
    {
        $var = [];
        $length = self::readLength($s);

        for ($i = 0; $i < $length; ++$i) {
            array_push($var, self::$read_func($s));
        }

        return $var;
    }

    public static function readStructList($s, $struct_name)
    {
        $var = [];
        $length = self::readLength($s);

        for ($i = 0; $i < $length; ++$i) {
            array_push($var, self::readStruct($s, $struct_name));
        }

        return $var;
    }

    public static function writeInt8($var)
    {
        return pack('C', $var);
    }

    public static function writeInt16($var)
    {
        return pack('n', $var);
    }

    public static function writeInt32($var)
    {
        return pack('N', $var);
    }

    public static function writeInt64($var)
    {
        return $var->encode();
    }

    public static function writeInt16V($var)
    {
        $var = self::convertToInt32($var) & 0xffff;

        if ($var < 255) {
            return self::writeInt8($var);
        } else {
            return self::writeInt8(255).self::writeInt16($var);
        }
    }

    public static function writeInt32V($var)
    {
        $var = self::convertToInt32($var);

        if ($var < 0) {
            return self::writeInt8(255).self::writeInt32($var);
        } else if ($var < 254) {
            return self::writeInt8($var);
        } else if ($var <= 65535) {
            return self::writeInt8(254).self::writeInt16($var);
        } else {
            return self::writeInt8(255).self::writeInt32($var);
        }
    }

    public static function writeInt64V($var)
    {
        $v = $var->toString();
        if (bccomp($v, '0') < 0) {
            $v = bcadd($v, '18446744073709551616');
        }

        if (bccomp($v, '253') < 0) {
            return self::writeInt8((int)$v);
        } else if (bccomp($v, '65535') <= 0) {
            return self::writeInt8(253).self::writeInt16((int)$v);
        } else if (bccomp($v, '4294967295') <= 0) {
            return self::writeInt8(254).self::writeInt32(self::convertToInt32($v));
        } else {
            return self::writeInt8(255).self::writeInt64($var);
        }
    }

    public static function writeInt16VZ($var)
    {
        return self::writeInt16V(self::zigzagEncode16($var));
    }

    public static function writeInt32VZ($var)
    {
        return self::writeInt32V(self::zigzagEncode32($var));
    }

    public static function writeInt64VZ($var)
    {
        return self::writeInt64V(self::zigzagEncode64($var));
    }

    public static function writeLength($var)
    {
        return self::writeInt32V($var);
    }

    public static function writeString($var)
    {
        return self::writeLength(strlen($var)).$var;
    }

    public static function writeStruct($var)
    {
        return $var->encode();
    }

    public static function writeList($var, $write_func)
    {
        $bin = self::writeLength(count($var));
        for ($i = 0; $i < count($var); ++$i) {
            $bin .= self::$write_func($var[$i]);
        }

        return $bin;
    }

    public static function readIntFromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        if (is_int($arr[$index]) ||
            is_float($arr[$index])) {
            return $arr[$index];
        }

        return (int)$arr[$index];
    }

    public static function readBoolFromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }

        return (bool)$arr[$index];
    }

    public static function readStringFromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }

        return (string)$arr[$index];
    }

    public static function readBytesFromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }

        return base64_decode($arr[$index]);
    }

    public static function readInt64FromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        return new \Brickred\Exchange\Int64((string)$arr[$index]);
    }

    public static function readUInt64FromArray($arr, $index)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        return new \Brickred\Exchange\UInt64((string)$arr[$index]);
    }

    public static function readStructFromArray($arr, $index, $struct_name)
    {
        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        if (!is_array($arr[$index])) {
            throw new CodecException("array['$index'] must be array");
        }

        $var = new $struct_name;
        $var->fromArray($arr[$index]);
        return $var;
    }

    public static function readListFromArray($arr, $index, $read_func)
    {
        $var = [];

        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        if (!is_array($arr[$index])) {
            throw new CodecException("array['$index'] must be array");
        }

        for ($i = 0; $i < count($arr[$index]); ++$i) {
            $var[$i] = self::$read_func($arr[$index], $i);
        }

        return $var;
    }

    public static function readStructListFromArray($arr, $index,
                                                   $struct_name)
    {
        $var = [];

        if (!isset($arr[$index])) {
            throw new CodecException("array['$index'] not set");
        }
        if (!is_array($arr[$index])) {
            throw new CodecException("array['$index'] must be array");
        }

        for ($i = 0; $i < count($arr[$index]); ++$i) {
            $var[$i] = self::readStructFromArray(
                $arr[$index], $i, $struct_name);
        }

        return $var;
    }

    private static function zigzagEncode16($var)
    {
        $var = self::convertToInt32($var) & 0xffff;
        if ($var > 32767) {
            $var -= 65536;
        }

        return (($var << 1) ^ ($var >> 15)) & 0xffff;
    }

    private static function zigzagEncode32($var)
    {
        $var = self::convertToInt32($var);
        $ret = ($var << 1) ^ ($var >> 31);

        if (PHP_INT_SIZE === 8) {
            return $ret;
        }

        if ($ret < 0) {
            $ret += 4294967296;
        }

        return $ret;
    }

    private static function zigzagDecode16($var)
    {
        $var = self::convertToInt32($var) & 0xffff;

        return ($var >> 1) ^ -($var & 1);
    }

    private static function zigzagDecode32($var)
    {
        $var  = self::convertToInt32($var);
        $half = ($var >> 1) & 0x7fffffff;
        return ($var & 1) ? -$half - 1 : $half;
    }

    private static function zigzagEncode64($var)
    {
        $high = $var->getHighInt32();
        $low  = $var->getLowInt32();

        // enc = (v << 1) ^ (v >> 63)
        // -- v << 1
        $carry    = ($low < 0) ? 1 : 0;
        $enc_high = self::convertToInt32($high * 2 + $carry);
        $enc_low  = self::convertToInt32($low * 2);
        // -- ^ all one equal ~
        if ($high < 0) {
            $enc_high = self::convertToInt32(~$enc_high);
            $enc_low  = self::convertToInt32(~$enc_low);
        }

        $ret = new UInt64();
        $ret->setHighInt32($enc_high);
        $ret->setLowInt32($enc_low);

        return $ret;
    }

    private static function zigzagDecode64($var)
    {
        $high = $var->getHighInt32();
        $low = $var->getLowInt32();

        // dec = (v >> 1) ^ -(v & 1)
        // -- v >> 1
        $half_high = ($high >> 1) & 0x7fffffff;
        $half_low  = self::convertToInt32(
            (($low >> 1) & 0x7fffffff) | (($high & 1) << 31));

        if (($low & 1) === 0) {
            // ^ all zero
            $dec_high = $half_high;
            $dec_low  = $half_low;
        } else {
            // ^ all one equal ~
            $dec_low  = self::convertToInt32(~$half_low);
            $dec_high = self::convertToInt32(~$half_high);
        }

        $ret = new Int64();
        $ret->setHighInt32($dec_high);
        $ret->setLowInt32($dec_low);

        return $ret;
    }
}
