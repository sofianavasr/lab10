CREATE TABLE clothes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    color VARCHAR(20) NOT NULL CHECK (color IN ('red', 'blue', 'pink', 'brown', 'black', 'white', 'gray', 'beige')),
    category VARCHAR(20) NOT NULL CHECK (category IN ('tops', 'bottoms', 'shoes')),
    style VARCHAR(20) NOT NULL CHECK (style IN ('casual', 'formal', 'business_casual', 'streetwear', 'athleisure')),
    weather VARCHAR(20) NOT NULL CHECK (weather IN ('cold', 'hot', 'rainy', 'snowy', 'windy', 'humid')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
