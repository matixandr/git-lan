use std::collections::HashMap;
use std::time::Instant;
use rand::Rng;

fn main() {
    let _currency = "PLN";
    let mut _balance = 100.0;
    
    let mut rarities: HashMap<&str, f32> = HashMap::new();
    let mut char_list = Vec::new();
    let mut index = 0;

    if rarities.len() < 8 {
        rarities.insert("♡", 0.9);
        rarities.insert("♢", 0.2);
        rarities.insert("♤", 0.01);
        rarities.insert("♧", 0.005);
        rarities.insert("♣", 0.001);
        rarities.insert("♦", 0.0005);
        rarities.insert("♥", 0.0001);
        rarities.insert("♠", 0.0000005);
        rarities.insert("✪", 0.0000001);
    }
    println!("\n\nStaring rarity roller...");
    let start_time = Instant::now();
    loop {
        index += 1;
        for _ in 0..10 {
            let mut rng = rand::thread_rng();
            let random_key = *rarities
                .iter()
                .max_by(|(_, &a), (_, &b)| a.partial_cmp(&b).unwrap())
                .map(|(key, _)| key)
                .unwrap();

            let chance = rng.gen::<f32>();
            let random_key = rarities
                .iter()
                .find(|(_, &value)| chance <= value)
                .map(|(key, _)| *key)
                .unwrap_or(random_key);
            char_list.push(random_key);
        }

        if char_list.contains(&"✪") {
            println!("\n\nWylosowano ✪ po {} próbach", index);
            break;
        } else {
            print!("\rPróba nr. {} z listą: {:?}            ",index,char_list);
            char_list.clear();
        }
    }
    let end_time = Instant::now();
    let duration = end_time - start_time;
    println!("Rolling took: {:?}\n\n", duration);
}
