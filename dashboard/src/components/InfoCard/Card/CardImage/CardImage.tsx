// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

export interface ICardImageProps {
  src: string;
  alt: string;
}

const CardImage = ({ src, alt }: ICardImageProps) => {
  return (
    <div className="card-img">
      <img src={src} alt={alt} />
    </div>
  );
};

export default CardImage;
