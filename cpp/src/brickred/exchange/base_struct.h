#ifndef BRICKRED_EXCHANGE_BASE_STRUCT_H
#define BRICKRED_EXCHANGE_BASE_STRUCT_H

#include <cstddef>
#include <cstdint>
#include <string>

namespace brickred::exchange {

class BaseStruct {
public:
    using CreateFunc = BaseStruct *(*)();

    BaseStruct();
    virtual ~BaseStruct();
    virtual BaseStruct *clone() const = 0;

    virtual int encode(char *buffer, size_t size) const = 0;
    virtual int decode(const char *buffer, size_t size) = 0;
    virtual std::string dump() const = 0;

protected:
    static std::string dumpBytes(const std::string &val);

protected:
    static inline uint16_t zigzagEncode16(int16_t v)  {
        return (uint16_t)(((uint32_t)v << 1) ^ (uint32_t)-(v < 0));
    }
    static inline uint32_t zigzagEncode32(int32_t v) {
        return ((uint32_t)v << 1) ^ (uint32_t)-(v < 0);
    }
    static inline uint64_t zigzagEncode64(int64_t v) {
        return ((uint64_t)v << 1) ^ (uint64_t)-(v < 0);
    }
    static inline int16_t zigzagDecode16(uint16_t v) {
        int16_t half = (int16_t)(v >> 1);
        return (v & 1u) ? (int16_t)(-half - 1) : half;
    }
    static inline int32_t zigzagDecode32(uint32_t v) {
        int32_t half = (int32_t)(v >> 1);
        return (v & 1u) ? -half - 1 : half;
    }
    static inline int64_t zigzagDecode64(uint64_t v) {
        int64_t half = (int64_t)(v >> 1);
        return (v & 1u) ? -half - 1 : half;
    }
};

} // namespace brickred::exchange

#endif
