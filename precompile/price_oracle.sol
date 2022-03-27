pragma solidity >=0.8.0;

interface NativePriceOracleInterface {
    
    function getPrice(uint256 identifier) external view returns (uint256);

    function getDecimals(uint256 identifier) external view returns (uint256);
}








